package plot_transformer

import (
	"fmt"
	"strconv"
)

// TransformRowsToFloat64 converts [][]string data to [][]float64 for plotting, setting invalid entries to zero.
func TransformRowsToFloat64(rows [][]string) [][]float64 {
	floatData := make([][]float64, len(rows))
	for i, row := range rows {
		floatData[i] = make([]float64, len(row))
		for j, val := range row {
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				f = 0
			}
			floatData[i][j] = f
		}
	}
	return floatData
}


// axisIndices finds the indices of the requested axes in headers.
func axisIndices(headers []string, xAxis, yAxis, zAxis string) (int, int, int) {
	xIndex, yIndex, zIndex := -1, -1, -1
	for i, h := range headers {
		if h == xAxis {
			xIndex = i
		}
		if h == yAxis {
			yIndex = i
		}
		if h == zAxis {
			zIndex = i
		}
	}
	return xIndex, yIndex, zIndex
}

// extractEventData extracts x/y/z data from event (string) rows.
func extractEventData(rows [][]string, xIndex, yIndex, zIndex int) (xData, yData, zData []interface{}) {
	for _, row := range rows {
		xData = append(xData, row[xIndex])
		yData = append(yData, row[yIndex])
		if zIndex >= 0 {
			zData = append(zData, row[zIndex])
		}
	}
	return
}

// extractFloatData extracts x/y/z data from numeric rows.
func extractFloatData(rows [][]string, xIndex, yIndex, zIndex int) (xData, yData, zData []interface{}) {
	floatData := TransformRowsToFloat64(rows)
	for _, row := range floatData {
		xData = append(xData, row[xIndex])
		yData = append(yData, row[yIndex])
		if zIndex >= 0 {
			zData = append(zData, row[zIndex])
		}
	}
	return
}

// TransformForPlot extracts x/y/z data and plot layout for plotting (e.g., Plotly) from headers and rows.
// If source == "events", x/y/z are string slices; otherwise, float64 slices.
func TransformForPlot(headers []string, rows [][]string, source, xAxis, yAxis, zAxis string) ([]map[string]interface{}, map[string]interface{}, error) {
	xIndex, yIndex, zIndex := axisIndices(headers, xAxis, yAxis, zAxis)
	if xIndex < 0 || yIndex < 0 || (zAxis != "" && zIndex < 0) {
		return nil, nil, fmt.Errorf("Invalid axes")
	}

	var xData, yData, zData []interface{}
	if source == "events" {
		xData, yData, zData = extractEventData(rows, xIndex, yIndex, zIndex)
	} else {
		xData, yData, zData = extractFloatData(rows, xIndex, yIndex, zIndex)
	}

	title := yAxis + " vs " + xAxis
	if zAxis != "" {
		title += " vs " + zAxis
	}

	plotLayout := map[string]interface{}{
		"title": title,
		"xaxis": map[string]string{"title": xAxis},
		"yaxis": map[string]string{"title": yAxis},
	}

	plotType := "scatter"
	if zAxis != "" {
		plotType = "scatter3d"
	}

	plotData := []map[string]interface{}{
		{
			"x":    xData,
			"y":    yData,
			"type": plotType,
			"mode": "markers",
		},
	}
	if zAxis != "" {
		plotData[0]["z"] = zData
	}

	return plotData, plotLayout, nil
}

