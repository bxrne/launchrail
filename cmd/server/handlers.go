package main

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/logger"
	"github.com/bxrne/launchrail/internal/plot_transformer"
	"github.com/bxrne/launchrail/internal/reporting"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/templates/pages"
	"github.com/gin-gonic/gin"
)

type DataHandler struct {
	records *storage.RecordManager
	Cfg     *config.Config
}

// Helper function to render templ components
func renderTempl(c *gin.Context, component templ.Component) {
	err := component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		err_err := c.AbortWithError(http.StatusInternalServerError, err)
		if err_err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render template"})
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
		renderTempl(c, pages.ErrorPage(err.Error()))
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

	renderTempl(c, pages.Data(pages.DataProps{
		Records: simRecords,
		Pagination: pages.Pagination{
			CurrentPage: params.Page,
			TotalPages:  totalPages,
		},
	}, h.Cfg.Setup.App.Version))
}

// DeleteRecord handles the request to delete a specific record
func (h *DataHandler) DeleteRecord(c *gin.Context) {
	hash := c.Param("hash")

	// Delete the record
	err := h.records.DeleteRecord(hash)
	if err != nil {
		renderTempl(c, pages.ErrorPage("Failed to delete record"))
		return
	}

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
		renderTempl(c, pages.ErrorPage(err.Error()))
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
	renderTempl(c, pages.RecordList(pages.DataProps{
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
		renderTempl(c, pages.ErrorPage("Invalid record identifier"))
		return
	}

	record, err := h.records.GetRecord(hash)
	if err != nil {
		renderTempl(c, pages.ErrorPage("Record not found"))
		return
	}
	defer record.Close()

	var store *storage.Storage
	var title string
	switch dataType {
	case "motion":
		store = record.Motion
		title = "Motion Data"
	case "events":
		store = record.Events
		title = "Events Data"
	case "dynamics":
		store = record.Dynamics
		title = "Dynamics Data"
	default:
		renderTempl(c, pages.ErrorPage("Invalid data type"))
		return
	}

	headers, data, err := store.ReadHeadersAndData()
	fmt.Println(headers, data, title)
	if err != nil {
		renderTempl(c, pages.ErrorPage("Failed to read data"))

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

	headers, data, err := storage.ReadHeadersAndData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read data"})
		return
	}

	// Sort the data
	// Find column index
	colIndex := -1
	for i, h := range headers {
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
	_, data, err := store.ReadHeadersAndData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read data"})
		return
	}

	itemsPerPage := 15 // Changed from 10 to 15
	startIndex := (page - 1) * itemsPerPage
	endIndex := min(startIndex+itemsPerPage, len(data))

	// Return only the table rows HTML
	c.HTML(http.StatusOK, "table_rows.html", gin.H{
		"rows": data[startIndex:endIndex],
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

// DownloadReport downloads a report for a specific record.
func (h *DataHandler) DownloadReport(c *gin.Context) {
	log := logger.GetLogger(h.Cfg.Setup.Logging.Level)
	recordID := c.Param("hash") // Use "hash" as per the route definition
	if recordID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Record hash/ID is required"})
		return
	}

	log.Info("Generating report", "recordID", recordID)

	// Assuming templates are relative to the executable or configured path
	// TODO: Make template path configurable or determine relative path reliably
	templateDir := "internal/reporting/templates" 
	reportGen, err := reporting.NewGenerator(templateDir)
	if err != nil {
		log.Error("Failed to create report generator", "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize report generator"})
		return
	}

	// --- Placeholder Data Loading & Plot Generation ---
	// In a real implementation, load data from storage based on recordID
	// and generate actual plots/maps.
	reportData, err := reporting.LoadSimulationData(recordID) // Placeholder load
	if err != nil {
		log.Warn("Using placeholder report data due to load error", "recordID", recordID, "loadError", err) 
		// Proceed with basic data if loading fails for now
		reportData = reporting.ReportData{
			RecordID: recordID,
			Version: h.Cfg.Setup.App.Version, // Get version from config
			// Add placeholder paths or indicate missing data
			GPSMapImagePath: "(Map generation not implemented)", 
		}
	} else {
		// Ensure version is set even if loading succeeds partially
		reportData.Version = h.Cfg.Setup.App.Version 
		// TODO: Add actual map generation logic here
		reportData.GPSMapImagePath = "(Map generation not implemented)" // Placeholder path
	}

	// --- Generate PDF ---
	pdfBytes, err := reportGen.GeneratePDF(reportData)
	if err != nil {
		// Log the detailed error
		log.Error("Failed to generate PDF report", "recordID", recordID, "error", err)
		
		// Check if it's the placeholder error vs. a real generation error
		if strings.Contains(err.Error(), "not yet implemented") || strings.Contains(err.Error(), "Placeholder") {
			// If it's just a placeholder issue, maybe return the markdown or a simpler error?
			// For now, return the placeholder PDF content which includes the markdown.
			log.Warn("PDF generation using placeholder", "recordID", recordID)
		} else {
			// If it's a different error, return a generic server error
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report"})
			return
		}
		// If we logged a warning but decided to proceed with placeholder PDF bytes, continue here.
	}

	if pdfBytes == nil {
	    log.Error("Generated PDF bytes are nil", "recordID", recordID)
	    c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report content"})
        return
	}


	// --- Send Response ---
	fileName := fmt.Sprintf("launch_report_%s.pdf", recordID)
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
	log.Info("Report sent successfully", "recordID", recordID, "fileName", fileName, "sizeBytes", len(pdfBytes))
}
