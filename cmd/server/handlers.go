package main

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/bxrne/launchrail/internal/storage"
	"github.com/bxrne/launchrail/templates/pages"
	"github.com/gin-gonic/gin"
)

type DataHandler struct {
	records *storage.RecordManager
}

func NewDataHandler(baseDir string) (*DataHandler, error) {
	rm, err := storage.NewRecordManager(baseDir)
	if err != nil {
		return nil, err
	}
	return &DataHandler{records: rm}, nil
}

// Helper function to render templ components
func renderTempl(c *gin.Context, component templ.Component) {
	err := component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
}

func (h *DataHandler) ListRecords(c *gin.Context) {
	records, err := h.records.ListRecords()
	if err != nil {
		renderTempl(c, pages.ErrorPage(err.Error()))
		return
	}

	// Convert storage.Record to pages.SimulationRecord
	simRecords := make([]pages.SimulationRecord, len(records))
	for i, record := range records {
		simRecords[i] = pages.SimulationRecord{
			Name:         record.Name,
			Hash:         record.Hash,
			LastModified: record.LastModified,
		}
	}

	renderTempl(c, pages.Data(pages.DataProps{Records: simRecords}))
}

// DeleteRecord handles the request to delete a specific record
func (h *DataHandler) DeleteRecord(c *gin.Context) {
	hash := c.Param("hash")
	err := h.records.DeleteRecord(hash)
	if err != nil {
		renderTempl(c, pages.ErrorPage("Failed to delete record"))
		return
	}

	c.Redirect(http.StatusFound, "/data")
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
