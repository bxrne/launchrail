package plot_transformer

import (
	"reflect"
	"testing"
)

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
