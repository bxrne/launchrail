package plot_transformer

import (
	"reflect"
	"testing"
)

func TestTransformForPlot_EventsSource(t *testing.T) {
	headers := []string{"time", "event", "value"}
	rows := [][]string{{"1", "ignite", "100"}, {"2", "burnout", "200"}}
	plotData, layout, err := TransformForPlot(headers, rows, "events", "time", "event", "value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plotData) == 0 || layout["title"] == "" {
		t.Error("Expected plot data and layout to be populated")
	}
}

func TestTransformForPlot_NumericSource(t *testing.T) {
	headers := []string{"x", "y", "z"}
	rows := [][]string{{"1.0", "2.0", "3.0"}, {"4.0", "5.0", "6.0"}}
	plotData, layout, err := TransformForPlot(headers, rows, "data", "x", "y", "z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plotData) == 0 || layout["title"] == "" {
		t.Error("Expected plot data and layout to be populated")
	}
}

func TestTransformForPlot_InvalidAxes(t *testing.T) {
	headers := []string{"a", "b"}
	rows := [][]string{{"1", "2"}}
	_, _, err := TransformForPlot(headers, rows, "data", "x", "y", "")
	if err == nil {
		t.Error("Expected error for invalid axes")
	}
}

func TestTransformRowsToFloat64(t *testing.T) {
	input := [][]string{
		{"1.2", "3.4"},
		{"5.6", "7.8"},
		{"bad", "0"},
	}
	want := [][]float64{
		{1.2, 3.4},
		{5.6, 7.8},
		{0, 0}, // 'bad' parses as 0
	}
	got := TransformRowsToFloat64(input)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
