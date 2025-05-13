package main

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/plot_transformer"
	"github.com/bxrne/launchrail/internal/reporting"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/internal/utils"

	"github.com/bxrne/launchrail/templates/pages"
	"github.com/gin-gonic/gin"
	"github.com/zerodha/logf"
)

var (
// TODO: Populate from viper config
// simulationDirName string = ".launchrail/simulations" // Removed unused variable
)

// HandlerRecordManager defines the subset of storage.RecordManager methods used by DataHandler.
type HandlerRecordManager interface {
	ListRecords() ([]*storage.Record, error)
	GetRecord(hash string) (*storage.Record, error)
	DeleteRecord(hash string) error
	GetStorageDir() string // Used by reporting.GenerateReportPackage
}

type DataHandler struct {
	records    HandlerRecordManager
	Cfg        *config.Config
	log        *logf.Logger
	ProjectDir string // Path to project root directory for finding templates
}

// NewDataHandler creates a new instance of DataHandler.
func NewDataHandler(records HandlerRecordManager, cfg *config.Config, log *logf.Logger) *DataHandler {
	// Try to determine the project directory for finding templates
	execPath, err := os.Executable()
	projectDir := ""
	if err == nil {
		// Use the executable's directory as a base and navigate up to find templates
		projectDir = filepath.Dir(execPath)
		// In production, the executable is likely in a binary directory, so we navigate up
		potentialTemplateDir := filepath.Join(projectDir, "templates")
		if _, err := os.Stat(potentialTemplateDir); os.IsNotExist(err) {
			// Go up one directory for development scenarios
			projectDir = filepath.Dir(projectDir)
		}
	}

	return &DataHandler{
		records:    records,
		Cfg:        cfg,
		log:        log,
		ProjectDir: projectDir,
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

	page, err := utils.ParseInt(c.Query("page"), "page")
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

	page, err := utils.ParseInt(c.Query("page"), "page")
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
	page, err := utils.ParseInt(c.Query("page"), "page")
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
	page, err := utils.ParseInt(c.Query("page"), "page")
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

	page, err := utils.ParseInt(c.Query("page"), "page")
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
	limit, err := utils.ParseInt(c.Query("limit"), "limit")
	if err != nil || limit < 0 { // Treat invalid or negative limit as no limit
		limit = 0
	}
	offset, err := utils.ParseInt(c.Query("offset"), "offset")
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ReportAPIV2 serves a specific report, potentially as a downloadable package or rendered view.
func (h *DataHandler) ReportAPIV2(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing report hash"})
		return
	}
	// Validate that the hash is a safe single path component
	if strings.Contains(hash, "/") || strings.Contains(hash, "\\") || strings.Contains(hash, "..") {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid report hash format"})
		return
	}

	format := c.DefaultQuery("format", "html") // Default to HTML, can be 'markdown', 'pdf', 'bundle'
	h.log.Info("Report requested", "hash", hash, "format", format)

	// Load the record data from storage
	record, err := h.records.GetRecord(hash)
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) { // Check for specific storage error
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		} else {
			h.log.Error("Failed to retrieve record from storage", "hash", hash, "error", err) // Log other errors
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to load report data"})
		}
		return
	}

	// Base directory for report assets (where SVGs, etc., are stored for this specific report)
	// assetsDir should point to something like ~/.launchrail/reports/HASH/assets/
	assetsDir := filepath.Join(record.Path, "assets")

	// The canonical path for report templates
	templatesDir := "/Users/adambyrne/code/fyp/launchrail/templates/reports/"

	// Validate if the canonical templates directory exists
	if _, statErr := os.Stat(templatesDir); os.IsNotExist(statErr) {
		h.log.Error("Canonical templates directory does not exist", "path", templatesDir, "error", statErr)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Server configuration error: templates directory not found"})
		return
	}
	h.log.Info("Using canonical templates directory", "path", templatesDir)

	var renderer *reporting.TemplateRenderer
	var rendererErr error

	// Initialize the template renderer
	renderer, rendererErr = reporting.NewTemplateRenderer(h.log, templatesDir, assetsDir)
	if rendererErr != nil {
		h.log.Error("Failed to initialize template renderer", "error", rendererErr)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize template renderer"})
		return
	}

	// Load simulation data for the report
	reportData, err := reporting.LoadSimulationData(hash, h.records, record.Path, h.Cfg)
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

	// First, check the Accept header for content negotiation
	acceptHeader := c.GetHeader("Accept")
	format = c.DefaultQuery("format", "html")

	// If Accept header is set to application/json, markdown or html explicitly, use that
	if strings.Contains(acceptHeader, "application/json") {
		format = "json"
	} else if strings.Contains(acceptHeader, "text/markdown") {
		format = "markdown"
	} else if strings.Contains(acceptHeader, "text/html") {
		format = "html"
	}

	switch format {
	case "json":
		// Return raw JSON data
		c.JSON(http.StatusOK, reportData)
	case "markdown", "md":
		// Render report as Markdown using the full FRR template
		reportContent, err := renderer.RenderToMarkdown(reportData, "report.md.tmpl")
		if err != nil {
			h.log.Error("Failed to render report", "error", err, "template", "report.md.tmpl")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to render report"})
			return
		}

		// Log successful rendering
		h.log.Info("Successfully rendered report with template", "template", "report.md.tmpl", "contentLength", len(reportContent))
		// Ensure proper line breaks are preserved in Markdown
		// Check if the report content already has proper line breaks
		if !strings.Contains(reportContent, "\n\n") {
			// Add proper line breaks to ensure Markdown formatting is preserved
			reportContent = strings.ReplaceAll(reportContent, "#", "\n#")
			reportContent = strings.ReplaceAll(reportContent, "* ", "\n* ")
			// Remove any extra line breaks that might have been introduced at the beginning
			reportContent = strings.TrimPrefix(reportContent, "\n")
		}

		// Save the report to the user's home directory in ~/.launchrail/reports/{hash}/index.md
		// Recreate the reportDir path to ensure it's correct
		currentUser, userErr := user.Current()
		if userErr != nil {
			h.log.Error("Failed to get current user for report directory", "error", userErr)
			// Skip report saving but still display the report
		} else {
			// Define paths explicitly to guarantee correctness
			reportDir := filepath.Join(currentUser.HomeDir, ".launchrail", "reports", hash)
			assetsDir := filepath.Join(reportDir, "assets")
			h.log.Info("Creating report directories", "reportDir", reportDir, "assetsDir", assetsDir)

			// Create fresh directory for reports
			// Remove if exists to start fresh with proper permissions
			if err := os.RemoveAll(reportDir); err != nil && !os.IsNotExist(err) {
				h.log.Warn("Error removing existing report directory", "path", reportDir, "error", err)
			}

			// Create the report and assets directories
			if err := os.MkdirAll(assetsDir, 0755); err != nil {
				h.log.Error("Failed to create report directories", "path", assetsDir, "error", err)
			} else {
				// Write the report file
				reportPath := filepath.Join(reportDir, "index.md")
				// Ensure we safely get a sample of the content for logging
				contentSample := ""
				if len(reportContent) > 0 {
					if len(reportContent) > 100 {
						contentSample = reportContent[:100] + "..."
					} else {
						contentSample = reportContent
					}
				}
				h.log.Info("Writing report to file", "path", reportPath, "content_length", len(reportContent), "content_sample", contentSample)

				// Write file contents with explicit error handling
				fileErr := os.WriteFile(reportPath, []byte(reportContent), 0644)
				if fileErr != nil {
					h.log.Error("Failed to write report file", "path", reportPath, "error", fileErr)
				} else {
					// Verify file was written successfully
					if _, statErr := os.Stat(reportPath); statErr != nil {
						h.log.Error("Report file could not be verified after writing", "path", reportPath, "error", statErr)
					} else {
						h.log.Info("Successfully saved and verified report file", "path", reportPath)

						// Generate placeholder SVG files for all plots
						for plotKey, plotPath := range reportData.Plots {
							// Create a simple placeholder SVG with more visible styling
							placeholderSVG := fmt.Sprintf(`<svg width="800" height="400" xmlns="http://www.w3.org/2000/svg">
<rect width="800" height="400" fill="#f0f0f0"/>
<rect width="780" height="380" x="10" y="10" fill="#e6e6e6" stroke="#999" stroke-width="2"/>
<text x="400" y="180" font-family="Arial" font-size="36" text-anchor="middle" fill="#333" font-weight="bold">%s</text>
<text x="400" y="240" font-family="Arial" font-size="20" text-anchor="middle" fill="#666">Placeholder Image</text>
</svg>`, plotKey)

							// Extract just the filename from the plotPath
							plotFilename := filepath.Base(plotPath)
							plotFullPath := filepath.Join(assetsDir, plotFilename)

							// Write the SVG file
							if writeErr := os.WriteFile(plotFullPath, []byte(placeholderSVG), 0644); writeErr != nil {
								h.log.Error("Failed to write plot file", "path", plotFullPath, "error", writeErr)
							} else {
								h.log.Info("Successfully saved plot file", "path", plotFullPath)
							}
						}
					}
				}
			}
		}

		c.Header("Content-Disposition", "inline; filename=report.md")
		c.Header("Content-Type", "text/markdown; charset=utf-8")
		c.String(http.StatusOK, reportContent)
	case "download":
		// Create a temporary directory for the report bundle
		tmpDir, err := os.MkdirTemp("", "report-*")
		if err != nil {
			h.log.Error("Failed to create temporary directory", "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare report for download"})
			return
		}
		defer os.RemoveAll(tmpDir)

		// Create the report bundle
		if err := renderer.CreateReportBundle(reportData, tmpDir); err != nil {
			h.log.Error("Failed to create report bundle", "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report bundle"})
			return
		}

		// Set filename for download
		filename := fmt.Sprintf("report-%s.md", hash[:8])
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Header("Content-Type", "text/markdown; charset=utf-8")

		// Read and serve the report file
		reportPath := filepath.Join(tmpDir, "report.md")
		c.File(reportPath)
	default: // html
		// Render report as Markdown first
		reportContent, err := renderer.RenderToMarkdown(reportData, "report.md.tmpl")
		if err != nil {
			h.log.Error("Failed to render report", "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to render report"})
			return
		}

		// Convert Markdown to HTML
		htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Simulation Report: %s</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            line-height: 1.6;
            color: #24292e;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        h1, h2, h3 { color: #0366d6; }
        table {
            border-collapse: collapse;
            width: 100%%;
            margin: 20px 0;
        }
        th, td {
            border: 1px solid #dfe2e5;
            padding: 8px 12px;
            text-align: left;
        }
        th { background-color: #f6f8fa; }
        img {
            max-width: 100%%;
            height: auto;
            display: block;
            margin: 20px 0;
        }
        pre {
            background-color: #f6f8fa;
            border-radius: 3px;
            padding: 16px;
            overflow: auto;
        }
        .report-header {
            border-bottom: 1px solid #eaecef;
            margin-bottom: 30px;
            padding-bottom: 10px;
        }
        .section {
            margin-bottom: 30px;
        }
    </style>
</head>
<body>
    <div class="report-header">
        <h1>Simulation Report: %s</h1>
        <p>Generated: %s</p>
    </div>
    <div class="report-content">
        %s
    </div>
</body>
</html>`, reportData.RecordID, reportData.RecordID, time.Now().Format(time.RFC1123), reportContent)

		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, htmlContent)
	}
}
