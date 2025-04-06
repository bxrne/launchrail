package main

import (
	"net/http"

	"github.com/bxrne/launchrail/internal/storage"
	"github.com/gin-gonic/gin"
)

type DataHandler struct {
	records *storage.RecordManager
}

type ViewData struct {
	Title   string
	Headers []string
	Data    [][]string
}

func NewDataHandler(baseDir string) (*DataHandler, error) {
	rm, err := storage.NewRecordManager(baseDir)
	if err != nil {
		return nil, err
	}
	return &DataHandler{records: rm}, nil
}

func (h *DataHandler) ListRecords(c *gin.Context) {
	records, err := h.records.ListRecords()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": err.Error(),
		})
		return
	}
	defer func() {
		for _, r := range records {
			r.Close()
		}
	}()

	c.HTML(http.StatusOK, "data.html", gin.H{
		"Records": records,
	})
}

func (h *DataHandler) GetRecordData(c *gin.Context) {
	hash := c.Param("hash")
	dataType := c.Param("type")

	record, err := h.records.GetRecord(hash)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Record not found",
		})
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
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Invalid data type",
		})
		return
	}

	data, err := store.ReadAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": err.Error(),
		})
		return
	}

	var headers []string
	var rows [][]string
	if len(data) > 0 {
		headers = data[0]
		rows = data[1:]
	}

	// Check if this is an HTMX request
	if c.GetHeader("HX-Request") == "true" {
		c.HTML(http.StatusOK, "data_view.html", gin.H{
			"Title":   title,
			"Headers": headers,
			"Data":    rows,
		})
		return
	}

	// Full page render if not HTMX
	c.HTML(http.StatusOK, "data.html", gin.H{
		"Records": []*storage.Record{record},
		"ViewData": ViewData{
			Title:   title,
			Headers: headers,
			Data:    rows,
		},
	})
}
