package main

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/plot_transformer"
	"github.com/bxrne/launchrail/internal/reporting"
	"github.com/bxrne/launchrail/internal/storage"

	"github.com/bxrne/launchrail/templates/pages"
	"github.com/gin-gonic/gin"
	"github.com/zerodha/logf"
)

// HandlerRecordManager defines the subset of storage.RecordManager methods used by DataHandler.
type HandlerRecordManager interface {
	ListRecords() ([]*storage.Record, error)
	GetRecord(hash string) (*storage.Record, error)
	DeleteRecord(hash string) error
	GetStorageDir() string // Used by reporting.GenerateReportPackage
}

type DataHandler struct {
	records HandlerRecordManager
	Cfg     *config.Config
	log     *logf.Logger
}

// NewDataHandler creates a new instance of DataHandler.
func NewDataHandler(records HandlerRecordManager, cfg *config.Config, log *logf.Logger) *DataHandler {
	return &DataHandler{
		records: records,
		Cfg:     cfg,
		log:     log,
	}
}

// Helper method to render templ components
// Accepts optional status codes. If provided and >= 400, sets the response status.
// Defaults to 200 OK otherwise.
func (h *DataHandler) renderTempl(c *gin.Context, component templ.Component, statusCodes ...int) {
	// Determine the status code to set
	statusCode := http.StatusOK // Default to 200 OK
	setStatus := false
	if len(statusCodes) > 0 && statusCodes[0] >= 400 {
		statusCode = statusCodes[0]
		setStatus = true // Mark that we intended to set a specific status
	}
	c.Status(statusCode) // Set the status code

	err := component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		h.log.Error("Failed to render template", "intended_status", statusCode, "error", err)
		// If we specifically set an error status code beforehand,
		// don't overwrite it with a 500 just because rendering failed.
		// Log the render error, but let the original status stand.
		if !setStatus && !c.Writer.Written() {
			// Only abort with 500 if we hadn't already set a specific error status
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to render template"})
		}
	}
}

type ListParams struct {
	Page         int
	ItemsPerPage int
}

func (h *DataHandler) ListRecords(c *gin.Context) {
	params := ListParams{
		Page:         1,
		ItemsPerPage: 15,
	}

	page, err := parseInt(c.Query("page"), "page")
	if err != nil || page < 1 {
		page = 1
	}
	params.Page = page

	records, err := h.records.ListRecords()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		h.renderTempl(c, pages.ErrorPage("Failed to list records"), http.StatusInternalServerError) // Pass status
		return
	}

	sortParam := c.Query("sort")
	if sortParam == "" {
		// default to newest first
		sortRecords(records, false)
	}

	// Calculate pagination
	totalRecords := len(records)
	totalPages := int(math.Ceil(float64(totalRecords) / float64(params.ItemsPerPage)))
	startIndex := (params.Page - 1) * params.ItemsPerPage
	endIndex := min(startIndex+params.ItemsPerPage, totalRecords)
	if startIndex >= totalRecords {
		startIndex = 0
		endIndex = min(params.ItemsPerPage, totalRecords)
		params.Page = 1
	}

	pagedRecords := records[startIndex:endIndex]

	// Convert to SimulationRecords
	simRecords := make([]pages.SimulationRecord, len(pagedRecords))
	for i, record := range pagedRecords {
		simRecords[i] = pages.SimulationRecord{
			Hash:         record.Hash,
			LastModified: record.LastModified,
		}
	}

	h.renderTempl(c, pages.Data(pages.DataProps{
		Records: simRecords,
		Pagination: pages.Pagination{
			CurrentPage: params.Page,
			TotalPages:  totalPages,
		},
	}, h.Cfg.Setup.App.Version)) // No status needed for success
}

// DeleteRecord handles the deletion of a specific simulation record.
func (h *DataHandler) DeleteRecord(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		h.log.Warn("DeleteRecord request missing hash")
		// Assume API request for this path
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Record hash is required"})
		return
	}
	h.log.Debug("Received API request to delete record", "hash", hash)

	// Delete the record
	err := h.records.DeleteRecord(hash)

	// Handle response based on error and request type
	if err != nil {
		h.log.Error("Failed to delete record", "hash", hash, "error", err)
		if errors.Is(err, storage.ErrRecordNotFound) { // Check for specific error
			h.log.Warn("Attempted to delete non-existent record", "hash", hash)
			// Assume API for this path
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		} else {
			// Generic internal server error for other deletion failures
			// Assume API for this path
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete record"})
		}
		return // Abort further processing
	}

	// Success
	h.log.Info("Successfully deleted record via API", "hash", hash)

	hxHeader := c.Request.Header.Get("Hx-Request")

	if hxHeader != "" {
		// HTMX request: Fetch updated records and render the list component
		updatedRecords, err := h.records.ListRecords()
		if err != nil {
			h.log.Error("Failed to list records after deletion for HTMX response", "error", err)
			// Send a generic error back to HTMX, or maybe an empty list with an error message?
			// For now, send 500
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Map storage.Record to pages.SimulationRecord
		pageRecords := make([]pages.SimulationRecord, 0, len(updatedRecords))
		for _, rec := range updatedRecords {
			pageRecords = append(pageRecords, pages.SimulationRecord{
				Hash:         rec.Hash,
				LastModified: rec.LastModified,
			})
		}

		// Prepare props for the RecordList component (no pagination for this simple swap)
		props := pages.DataProps{
			Records: pageRecords,
			// Pagination: pages.Pagination{}, // Omit pagination for now
		}

		// Set content type and render the component
		c.Header("Content-Type", "text/html; charset=utf-8")
		err = pages.RecordList(props).Render(c.Request.Context(), c.Writer)
		if err != nil {
			h.log.Error("Failed to render RecordList component for HTMX response", "error", err)
			// Abort if rendering fails, status code is already set potentially by Render
			return
		}
		// Status OK is implicit on successful render without Abort

	} else {
		// API request: Respond with No Content
		c.AbortWithStatus(http.StatusNoContent)
	}
}

// DeleteRecord handles the request to delete a specific record
func (h *DataHandler) DeleteRecordOld(c *gin.Context) {
	hash := c.Param("hash")

	// Delete the record
	err := h.records.DeleteRecord(hash)
	if err != nil {
		// Handle errors first, regardless of request type
		if strings.Contains(err.Error(), "not found") {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "record not found"})
		} else {
			// Log the unexpected error
			h.log.Error("Failed to delete record", "error", err, "recordID", hash)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to delete record"})
		}
		return
	}

	// Deletion successful, now determine response type
	if !strings.Contains(c.GetHeader("Accept"), "text/html") {
		// API request: Return No Content and stop processing
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	// --- HTMX Request: Proceed to render the updated list ---

	// Prepare pagination parameters
	params := ListParams{
		Page:         1,
		ItemsPerPage: 15,
	}

	page, err := parseInt(c.Query("page"), "page")
	if err != nil || page < 1 {
		page = 1
	}
	params.Page = page

	// Retrieve and (re)sort records
	records, err := h.records.ListRecords()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		h.renderTempl(c, pages.ErrorPage("Failed to list records"), http.StatusInternalServerError) // Pass status
		return
	}

	sortParam := c.Query("sort")
	if sortParam == "" {
		sortRecords(records, false) // newest first by default
	} else {
		sortRecords(records, sortParam == "time_asc")
	}

	// Calculate pagination window
	totalRecords := len(records)
	totalPages := int(math.Ceil(float64(totalRecords) / float64(params.ItemsPerPage)))
	startIndex := (params.Page - 1) * params.ItemsPerPage
	endIndex := min(startIndex+params.ItemsPerPage, totalRecords)
	if startIndex >= totalRecords {
		startIndex = 0
		endIndex = min(params.ItemsPerPage, totalRecords)
		params.Page = 1
	}

	pagedRecords := records[startIndex:endIndex]

	// Convert to SimulationRecords
	simRecords := make([]pages.SimulationRecord, len(pagedRecords))
	for i, record := range pagedRecords {
		simRecords[i] = pages.SimulationRecord{
			Hash:         record.Hash,
			LastModified: record.LastModified,
		}
	}

	// Render only the updated record list (partial HTML) so htmx can swap it in-place
	h.renderTempl(c, pages.RecordList(pages.DataProps{
		Records: simRecords,
		Pagination: pages.Pagination{
			CurrentPage: params.Page,
			TotalPages:  totalPages,
		},
	}))
}

// GetRecordData handles the request to get data from a specific record
func (h *DataHandler) GetRecordData(c *gin.Context) {
	hash := c.Param("hash")
	dataType := c.Param("type")

	// Validate the hash to ensure it is a single-component identifier
	if strings.Contains(hash, "/") || strings.Contains(hash, "\\") || strings.Contains(hash, "..") {
		h.renderTempl(c, pages.ErrorPage("Invalid record identifier"), http.StatusBadRequest) // Pass status
		return
	}

	record, err := h.records.GetRecord(hash)
	if err != nil {
		h.renderTempl(c, pages.ErrorPage("Record not found"), http.StatusNotFound) // Pass status
		return
	}
	defer record.Close()

	var store *storage.Storage
	switch dataType {
	case "motion":
		store = record.Motion
	case "events":
		store = record.Events
	case "dynamics":
		store = record.Dynamics
	default:
		h.renderTempl(c, pages.ErrorPage("Invalid data type"), http.StatusBadRequest) // Pass status
		return
	}

	_, _, err = store.ReadHeadersAndData() // Use '=' as err is already declared
	if err != nil {
		h.renderTempl(c, pages.ErrorPage("Failed to read data"), http.StatusInternalServerError) // Pass status
		return
	}

	// For now, redirect to the explore page which will show the data
	// You might want to create a specific templ component for data plots later
	c.Redirect(http.StatusFound, fmt.Sprintf("/explore/%s", hash))
}

// New endpoint to return JSON data for explorer
func (h *DataHandler) GetExplorerData(c *gin.Context) {
	hash := c.Param("hash")
	record, err := h.records.GetRecord(hash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}
	defer record.Close()

	motionHeaders, motionData, err := record.Motion.ReadHeadersAndData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read motion data"})
		return
	}
	dynamicsHeaders, dynamicsData, err := record.Dynamics.ReadHeadersAndData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read dynamics data"})
		return
	}
	eventsHeaders, eventsData, err := record.Events.ReadHeadersAndData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read events data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"headers": gin.H{
			"motion":   motionHeaders,
			"dynamics": dynamicsHeaders,
			"events":   eventsHeaders,
		},
		"data": gin.H{
			"motion":   motionData,
			"dynamics": dynamicsData,
			"events":   eventsData,
		},
	})
}

// Add this new method to DataHandler
func (h *DataHandler) ExplorerSortData(c *gin.Context) {
	hash := c.Param("hash")
	table := c.Query("table")
	column := c.Query("col")
	direction := c.Query("dir")
	page, err := parseInt(c.Query("page"), "page")
	if err != nil || page < 1 {
		page = 1
	}

	record, err := h.records.GetRecord(hash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}
	defer record.Close()

	// Get the correct storage based on table type
	var storage *storage.Storage
	switch table {
	case "motion":
		storage = record.Motion
	case "dynamics":
		storage = record.Dynamics
	case "events":
		storage = record.Events
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table"})
		return
	}

	_, data, err := storage.ReadHeadersAndData() // Use blank identifier for headers as it's not directly used here
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read data"})
		return
	}

	// Sort the data
	// Find column index
	colIndex := -1
	var sortHeaders []string
	if len(data) > 0 {
		// Assuming the first row of data contains headers if headers aren't fetched separately
		// This needs verification based on how storage.ReadHeadersAndData works.
		// Let's assume for now we need to fetch headers properly if they aren't in data[0]
		headers, _, err := storage.ReadHeadersAndData() // Fetch headers specifically for sorting
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read headers for sorting"})
			return
		}
		sortHeaders = headers
	} else {
		// Handle case with no data? Or assume headers can still be fetched?
		// For now, use an empty slice if no data.
		sortHeaders = []string{}
	}

	for i, h := range sortHeaders { // Correct: Iterate over actual headers fetched for sorting
		if h == column {
			colIndex = i
			break
		}
	}

	// Convert and sort data
	sortedData := make([][]string, len(data))
	copy(sortedData, data)

	sort.Slice(sortedData, func(i, j int) bool {
		if direction == "asc" {
			return sortedData[i][colIndex] < sortedData[j][colIndex]
		}
		return sortedData[i][colIndex] > sortedData[j][colIndex]
	})

	// Apply pagination
	itemsPerPage := 15 // Changed from 10 to 15
	totalPages := int(math.Ceil(float64(len(sortedData)) / float64(itemsPerPage)))
	startIndex := (page - 1) * itemsPerPage
	endIndex := min(startIndex+itemsPerPage, len(sortedData))
	pagedData := sortedData[startIndex:endIndex]

	// Return paginated and sorted data
	c.JSON(http.StatusOK, gin.H{
		"data": pagedData,
		"pagination": gin.H{
			"currentPage": page,
			"totalPages":  totalPages,
		},
	})
}

func (h *DataHandler) GetTableRows(c *gin.Context) {
	hash := c.Param("hash")
	table := c.Query("table")
	page, err := parseInt(c.Query("page"), "page")
	if err != nil || page < 1 {
		page = 1
	}

	record, err := h.records.GetRecord(hash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}
	defer record.Close()

	var store *storage.Storage
	switch table {
	case "motion":
		store = record.Motion
	case "dynamics":
		store = record.Dynamics
	case "events":
		store = record.Events
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table"})
		return
	}

	// Get data and paginate
	headers, data, err := store.ReadHeadersAndData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read data"})
		return
	}

	itemsPerPage := 15
	startIndex := (page - 1) * itemsPerPage
	endIndex := min(startIndex+itemsPerPage, len(data))

	// Return only the table rows HTML
	c.HTML(http.StatusOK, "table_rows.html", gin.H{
		"headers": headers,
		"rows":    data[startIndex:endIndex],
	})
}

func (h *DataHandler) handleTableRequest(c *gin.Context, hash string, table string) {
	record, err := h.records.GetRecord(hash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}
	defer record.Close()

	var store *storage.Storage
	switch table {
	case "motion":
		store = record.Motion
	case "dynamics":
		store = record.Dynamics
	case "events":
		store = record.Events
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table"})
		return
	}

	page, err := parseInt(c.Query("page"), "page")
	if err != nil || page < 1 {
		page = 1
	}
	sortCol := c.Query("sort")
	sortDir := c.Query("dir")

	headers, data, err := store.ReadHeadersAndData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read data"})
		return
	}

	// Sort if requested
	if sortCol != "" {
		colIndex := -1
		for i, h := range headers {
			if h == sortCol {
				colIndex = i
				break
			}
		}
		if colIndex >= 0 {
			sort.Slice(data, func(i, j int) bool {
				if sortDir == "asc" {
					return data[i][colIndex] < data[j][colIndex]
				}
				return data[i][colIndex] > data[j][colIndex]
			})
		}
	}

	// Paginate the data
	itemsPerPage := 15
	totalPages := int(math.Ceil(float64(len(data)) / float64(itemsPerPage)))
	startIndex := (page - 1) * itemsPerPage
	endIndex := min(startIndex+itemsPerPage, len(data))
	if startIndex >= len(data) {
		startIndex = 0
		endIndex = min(itemsPerPage, len(data))
		page = 1
	}

	pagedData := data[startIndex:endIndex]

	// For motion and dynamics, convert string data to float64
	if table != "events" {
		floatData := plot_transformer.TransformRowsToFloat64(pagedData)
		c.JSON(http.StatusOK, gin.H{
			"headers": headers,
			"data":    floatData,
			"pagination": gin.H{
				"currentPage": page,
				"totalPages":  totalPages,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"headers": headers,
		"data":    pagedData,
		"pagination": gin.H{
			"currentPage": page,
			"totalPages":  totalPages,
		},
	})
}

// Use this standard handler for all table requests
func (h *DataHandler) GetTableData(c *gin.Context) {
	hash := c.Param("hash")
	table := c.Query("table")
	h.handleTableRequest(c, hash, table)
}

// sortRecords sorts records based on CreationTime timestamp
func sortRecords(records []*storage.Record, ascending bool) {
	sort.Slice(records, func(i, j int) bool {
		if ascending {
			return records[i].CreationTime.Before(records[j].CreationTime)
		}
		return records[i].CreationTime.After(records[j].CreationTime)
	})
}

// ListRecordsAPI godoc
// @Summary List simulation records
// @Description Returns a paginated list of simulation records
// @Tags Data
// @Accept json
// @Produce json
// @Param filter query string false "Filter by hash"
// @Param sort query string false "Sort order (time_asc or time_desc)"
// @Param limit query int false "Limit the number of records"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} RecordsResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/data [get]
func (h *DataHandler) ListRecordsAPI(c *gin.Context) {
	records, err := h.records.ListRecords()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Apply query parameters
	filter := c.Query("filter")
	sortOrder := c.Query("sort")

	// Parse limit and offset for pagination
	limit, err := parseInt(c.Query("limit"), "limit")
	if err != nil || limit < 0 { // Treat invalid or negative limit as no limit
		limit = 0
	}
	offset, err := parseInt(c.Query("offset"), "offset")
	if err != nil || offset < 0 {
		offset = 0
	}

	// Apply filtering *before* calculating total for pagination
	if filter != "" {
		filtered := make([]*storage.Record, 0)
		for _, r := range records {
			if strings.Contains(strings.ToLower(r.Hash), strings.ToLower(filter)) {
				filtered = append(filtered, r)
			}
		}
		records = filtered
	}

	sortRecords(records, sortOrder == "time_asc")

	// Calculate pagination based on limit and offset
	totalRecords := len(records)
	startIndex := offset
	endIndex := totalRecords // Default endIndex is the total number of records

	// Apply limit if it's greater than 0
	if limit > 0 {
		endIndex = min(startIndex+limit, totalRecords)
	}

	// Ensure startIndex is within bounds
	if startIndex >= totalRecords {
		// If offset is past the end, return an empty list but correct total
		startIndex = totalRecords
		endIndex = totalRecords
	} else if startIndex < 0 { // Should not happen with default 0, but good practice
		startIndex = 0
	}

	// Return paginated records with total count
	c.JSON(http.StatusOK, gin.H{
		"total":   totalRecords, // Add total count
		"records": records[startIndex:endIndex],
	})
}

func parseFloat(valueStr string, fieldName string) (float64, error) {
	if valueStr == "" {
		return 0, fmt.Errorf("%s is required", fieldName)
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", fieldName, err)
	}
	return value, nil
}

func parseInt(valueStr string, fieldName string) (int, error) {
	if valueStr == "" {
		return 0, fmt.Errorf("%s is required", fieldName)
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", fieldName, err)
	}
	return value, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ReportAPIV2 serves a specific report, potentially as a downloadable package or rendered view.
// TODO: This currently returns JSON data. Needs to be adapted for actual report serving (HTML/Zip).
func (h *DataHandler) ReportAPIV2(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing report hash"})
		return
	}
	// Validate that the hash is a safe single path component
	if strings.Contains(hash, "/") || strings.Contains(hash, "\\") || strings.Contains(hash, "..") {
		h.log.Warn("Invalid report hash provided", "hash", hash)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid report hash"})
		return
	}
	h.log.Info("Report data requested", "hash", hash)

	// Use the RecordManager's configured storage directory
	baseRecordsDir := h.records.GetStorageDir()
	if baseRecordsDir == "" {
		h.log.Error("Base records directory is not configured or accessible in RecordManager")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error: Records directory configuration error"})
		return
	}
	reportSpecificDir := filepath.Join(baseRecordsDir, hash)
	// Ensure the resolved path is within the base directory
	absReportDir, err := filepath.Abs(reportSpecificDir)
	if err != nil || !strings.HasPrefix(absReportDir, filepath.Clean(baseRecordsDir)+string(os.PathSeparator)) {
		h.log.Warn("Resolved report directory is outside the base directory", "resolvedDir", absReportDir, "baseDir", baseRecordsDir)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid report hash"})
		return
	}

	// Ensure h.records is not nil and is of type *storage.RecordManager
	rm, ok := h.records.(*storage.RecordManager)
	if !ok || rm == nil {
		h.log.Error("RecordManager is not initialized or of incorrect type in handler")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error: Record manager not available"})
		return
	}

	reportData, err := reporting.LoadSimulationData(hash, rm, reportSpecificDir, h.Cfg)
	if err != nil {
		h.log.Error("Failed to load simulation data for report", "hash", hash, "error", err)
		if errors.Is(err, storage.ErrRecordNotFound) ||
			strings.Contains(strings.ToLower(err.Error()), "record not found") ||
			strings.Contains(err.Error(), "no such file or directory") || // Check for file system errors too
			strings.Contains(err.Error(), "failed to get record") { // Check for our specific GetRecord error
			h.log.Warn("Report requested for non-existent record or data loading failed", "hash", hash)
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Data for report not found or incomplete"})
		} else {
			// For other errors, return a generic server error or a more specific one if appropriate
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report data"})
		}
		return
	}

	c.JSON(http.StatusOK, reportData)
}

// CreateRecordAPI handles the creation of a new simulation record.
