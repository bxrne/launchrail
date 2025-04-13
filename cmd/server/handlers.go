package main

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"

	"github.com/a-h/templ"
	"github.com/bxrne/launchrail/internal/config"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/templates/pages"
	"github.com/gin-gonic/gin"
)

type DataHandler struct {
	records *storage.RecordManager
	Cfg     *config.Config
}

func NewDataHandler(cfg *config.Config) (*DataHandler, error) {
	rm, err := storage.NewRecordManager(cfg.Setup.App.BaseDir)
	if err != nil {
		return nil, err
	}
	return &DataHandler{records: rm, Cfg: cfg}, nil
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
	Sort         string
	Filter       string
	ItemsPerPage int
}

func (h *DataHandler) ListRecords(c *gin.Context) {
	params := ListParams{
		Page:         parseInt(c.Query("page"), 1),
		Sort:         c.Query("sort"),
		Filter:       c.Query("filter"),
		ItemsPerPage: 10,
	}

	records, err := h.records.ListRecords()
	if err != nil {
		renderTempl(c, pages.ErrorPage(err.Error()))
		return
	}

	// Apply filtering
	if params.Filter != "" {
		filtered := make([]*storage.Record, 0)
		for _, r := range records {
			if strings.Contains(strings.ToLower(r.Hash), strings.ToLower(params.Filter)) {
				filtered = append(filtered, r)
			}
		}
		records = filtered
	}

	// Apply sorting
	sort.Slice(records, func(i, j int) bool {
		switch params.Sort {
		case "time_asc":
			return records[i].LastModified.Before(records[j].LastModified)
		default: // time_desc
			return records[i].LastModified.After(records[j].LastModified)
		}
	})

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
	},
		h.Cfg.Setup.App.Version,
	))
}

// DeleteRecord handles the request to delete a specific record
func (h *DataHandler) DeleteRecord(c *gin.Context) {
	hash := c.Param("hash")
	err := h.records.DeleteRecord(hash)
	if err != nil {
		renderTempl(c, pages.ErrorPage("Failed to delete record"))
		return
	}

	// Get updated records list
	records, err := h.records.ListRecords()
	if err != nil {
		renderTempl(c, pages.ErrorPage(err.Error()))
		return
	}

	// Convert to SimulationRecord slice
	simRecords := make([]pages.SimulationRecord, len(records))
	for i, record := range records {
		simRecords[i] = pages.SimulationRecord{
			Name:         record.Name,
			Hash:         record.Hash,
			LastModified: record.LastModified,
		}
	}

	// Render just the records list component
	renderTempl(c, pages.Data(pages.DataProps{Records: simRecords}, h.Cfg.Setup.App.Version))
}

// GetRecordData handles the request to get data from a specific record
func (h *DataHandler) GetRecordData(c *gin.Context) {
	hash := c.Param("hash")
	dataType := c.Param("type")

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
	page := parseInt(c.Query("page"), 1)

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
	itemsPerPage := 10
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

// New helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (h *DataHandler) GetTableRows(c *gin.Context) {
	hash := c.Param("hash")
	table := c.Query("table")
	page := parseInt(c.Query("page"), 1)

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

	itemsPerPage := 10
	startIndex := (page - 1) * itemsPerPage
	endIndex := min(startIndex+itemsPerPage, len(data))

	// Return only the table rows HTML
	c.HTML(http.StatusOK, "table_rows.html", gin.H{
		"rows": data[startIndex:endIndex],
	})
}
