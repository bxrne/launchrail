package main

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
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

	// We'll use the HandlerRecordManager interface directly - no need for type assertion

	// Create assets directory if it doesn't exist
	assetsDir := filepath.Join(reportSpecificDir, "assets")
	if err := os.MkdirAll(assetsDir, os.ModePerm); err != nil {
		h.log.Error("Failed to create assets directory", "path", assetsDir, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create assets directory"})
		return
	}

	// Load simulation data
	reportData, err := reporting.LoadSimulationData(hash, h.records, reportSpecificDir, h.Cfg)
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

	// Check if the client specifically wants JSON (for API compatibility)
	if c.GetHeader("Accept") == "application/json" {
		c.JSON(http.StatusOK, reportData)
		return
	}

	// Otherwise, render a proper report using our template renderer
	// Enhanced debugging for template directory resolution
	h.log.Debug("Starting template directory lookup process")

	// Initialize templates and declare variables
	var (
		templatesDir string
		renderer     *reporting.TemplateRenderer
		absPath      string
		tempErr      error // Temporary error variable for operations
	)

	// Try multiple approaches to find the templates directory
	templateFound := false

	// First approach: Check in cmd/server/testdata directory specifically (for test environment)
	if !templateFound {
		// Get the current working directory for absolute path construction
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			cwd = "" // Reset if we can't get it
		}

		// Create a comprehensive list of possible template locations
		possibleTestPaths := []string{
			// Absolute paths - these will be most reliable
			filepath.Join(cwd, "cmd", "server", "testdata", "templates", "reports"),
			filepath.Join(cwd, "testdata", "templates", "reports"),

			// Relative paths - for when the working directory is already at the right level
			filepath.Join("cmd", "server", "testdata", "templates", "reports"),
			filepath.Join("testdata", "templates", "reports"),
			filepath.Join("templates", "reports"),

			// Specific path we know exists from examining the file system
			"/Users/adambyrne/code/fyp/launchrail/cmd/server/testdata/templates/reports",
		}

		for _, testPath := range possibleTestPaths {
			absPath, tempErr = filepath.Abs(testPath)
			if tempErr == nil {
				h.log.Debug("Checking test template directory", "path", absPath)
				if _, err := os.Stat(testPath); err == nil {
					// We found a valid test template directory
					templatesDir = testPath
					h.log.Info("Found valid test templates directory", "path", absPath)
					templateFound = true
					break
				}
			}
		}
	}

	// Second approach: Try looking for templates relative to the working directory
	if !templateFound {
		var cwd string
		cwd, err = os.Getwd()
		if err == nil {
			h.log.Debug("Current working directory", "cwd", cwd)
		}

		// Check if templates exist directly under the working directory
		directTemplatesDir := filepath.Join("templates", "reports")
		if _, err := os.Stat(directTemplatesDir); err == nil {
			h.log.Debug("Found templates directly in working directory", "path", directTemplatesDir)
			// Use this directory for templates
			templatesDir = directTemplatesDir
			absPath, _ = filepath.Abs(templatesDir)
			h.log.Info("Using templates from working directory", "path", absPath)
			templateFound = true
		}
	}

	// Next approach: Try ProjectDir-based path
	// templatesDir is already declared above
	if h.ProjectDir != "" {
		h.log.Debug("Checking ProjectDir for templates", "ProjectDir", h.ProjectDir)
		templatesDir = filepath.Join(h.ProjectDir, "templates", "reports")
		absPath, _ := filepath.Abs(templatesDir)
		h.log.Debug("Constructed templates path", "path", templatesDir, "abs_path", absPath)

		// Verify that the directory exists
		if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
			h.log.Warn("Templates directory not found in ProjectDir, will try fallbacks", "dir", templatesDir)
			templatesDir = "" // Clear to fall back to alternate method
		} else {
			h.log.Info("Using templates from ProjectDir", "dir", templatesDir)
		}
	} else {
		h.log.Debug("ProjectDir is empty, skipping this lookup method")
	}

	// Fall back to executable directory if ProjectDir didn't work
	if templatesDir == "" {
		h.log.Debug("Using executable path fallback method")
		execDir, err := os.Executable()
		if err != nil {
			h.log.Error("Failed to get executable path", "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to determine templates directory"})
			return
		}
		templatesDir = filepath.Join(filepath.Dir(execDir), "templates", "reports")
		absPath, _ := filepath.Abs(templatesDir)
		h.log.Debug("Constructed templates path from executable", "exec_dir", filepath.Dir(execDir), "templates_dir", templatesDir, "abs_path", absPath)

		// Check if this directory exists
		if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
			h.log.Warn("Templates directory not found from executable path", "dir", templatesDir)

			// Last resort: Try to find templates in the project root
			repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(execDir))) // Go up 3 levels
			templatesDir = filepath.Join(repoRoot, "templates", "reports")
			h.log.Debug("Last resort template path", "dir", templatesDir)

			if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
				h.log.Error("Failed to find templates directory through all methods", "last_attempt", templatesDir)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not locate templates directory"})
				return
			}
		}
	}

	// Create the template renderer once we've found a valid templates directory
	if templateFound && templatesDir != "" {
		var rendererErr error
		renderer, rendererErr = reporting.NewTemplateRenderer(h.log, templatesDir, assetsDir)
		if rendererErr != nil {
			h.log.Error("Failed to create template renderer", "error", rendererErr, "templatesDir", templatesDir)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize report renderer"})
			return
		}
	} else {
		// We couldn't find a valid templates directory
		h.log.Error("Could not find any templates directory after trying all methods")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not locate templates directory"})
		return
	}

	// First, check the Accept header for content negotiation
	acceptHeader := c.GetHeader("Accept")
	format := c.DefaultQuery("format", "html")

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
		// Render report as Markdown
		reportContent, err := renderer.RenderReport(reportData)
		if err != nil {
			h.log.Error("Failed to render report", "error", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to render report"})
			return
		}
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
		reportContent, err := renderer.RenderReport(reportData)
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

// CreateRecordAPI handles the creation of a new simulation record.
