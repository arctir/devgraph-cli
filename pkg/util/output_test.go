package util

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisplayTable(t *testing.T) {
	tests := []struct {
		name     string
		data     []map[string]interface{}
		headers  []string
		contains []string
	}{
		{
			name: "simple table",
			data: []map[string]interface{}{
				{"name": "Alice", "age": 30, "city": "New York"},
				{"name": "Bob", "age": 25, "city": "San Francisco"},
			},
			headers:  []string{"name", "age", "city"},
			contains: []string{"Alice", "30", "New York", "Bob", "25", "San Francisco"},
		},
		{
			name: "mixed data types",
			data: []map[string]interface{}{
				{"id": 1, "score": 95.5, "active": true},
				{"id": 2, "score": 87.25, "active": false},
			},
			headers:  []string{"id", "score", "active"},
			contains: []string{"1", "95.50", "true", "2", "87.25", "false"},
		},
		{
			name: "missing values",
			data: []map[string]interface{}{
				{"name": "Alice", "age": 30},
				{"name": "Bob", "city": "Boston"},
			},
			headers:  []string{"name", "age", "city"},
			contains: []string{"Alice", "30", "Bob", "Boston"},
		},
		{
			name:     "empty data",
			data:     []map[string]interface{}{},
			headers:  []string{"name", "age"},
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			DisplayTable(tt.data, tt.headers)

			err := w.Close()
			if err != nil {
				t.Fatalf("Failed to close pipe: %v", err)
			}
			os.Stdout = old

			var buf bytes.Buffer
			_, err = buf.ReadFrom(r)
			if err != nil {
				t.Fatalf("Failed to read from pipe: %v", err)
			}
			output := buf.String()

			for _, expected := range tt.contains {
				if expected != "" {
					assert.Contains(t, output, expected, "Expected output to contain: %s", expected)
				}
			}

			for _, header := range tt.headers {
				assert.Contains(t, strings.ToUpper(output), strings.ToUpper(header),
					"Expected output to contain header: %s", header)
			}
		})
	}
}

func TestDisplayTable_EmptyHeaders(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := []map[string]interface{}{
		{"name": "Alice", "age": 30},
	}
	headers := []string{}

	DisplayTable(data, headers)

	err := w.Close()
	if err != nil {
		t.Fatalf("Failed to close pipe: %v", err)
	}
	os.Stdout = old

	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	assert.NotContains(t, output, "Alice")
	assert.NotContains(t, output, "30")
}
