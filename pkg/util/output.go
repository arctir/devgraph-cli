// Package util provides utility functions for the Devgraph CLI.
// It includes functions for output formatting, table display, and common operations
// used across multiple commands.
package util

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

// DisplayTable takes a slice of maps (data) and headers, and displays it as a formatted table.
// Each map represents a row of data, with keys corresponding to column headers.
// The function handles different data types (string, int, float64) and formats them appropriately.
// Missing values are displayed as "-" in a dimmed color.
func DisplayTable(data []map[string]interface{}, headers []string) {
	// Add some visual spacing before the table
	fmt.Println()

	// Create a new table writer
	table := tablewriter.NewWriter(os.Stdout)

	// Style the headers with colors but keep original case
	styledHeaders := make([]string, len(headers))
	for i, header := range headers {
		styledHeaders[i] = color.New(color.FgBlue, color.Bold).Sprint(header)
	}
	table.Header(styledHeaders)

	// Convert map data to table rows
	for _, row := range data {
		var rowData []string
		for _, header := range headers {
			// Convert interface{} to string
			var value string
			if val, ok := row[header]; ok {
				switch v := val.(type) {
				case string:
					value = truncateString(v, 60) // Limit very long strings
				case int:
					value = fmt.Sprintf("%d", v)
				case float64:
					value = fmt.Sprintf("%.2f", v)
				default:
					value = truncateString(fmt.Sprintf("%v", v), 60)
				}
			} else {
				value = color.New(color.FgHiBlack).Sprint("-")
			}
			rowData = append(rowData, value)
		}
		if err := table.Append(rowData); err != nil {
			log.Printf("Warning: Failed to append table row: %v", err)
		}
	}

	// The tablewriter package has limited customization options available
	// Focus on content improvements instead of styling

	// Render the table
	if err := table.Render(); err != nil {
		log.Printf("Warning: Failed to render table: %v", err)
	}

	// Add some spacing after the table
	fmt.Println()
}

// truncateString truncates a string to a maximum length with ellipsis.
// If the string is shorter than or equal to maxLen, it returns the original string.
// If maxLen is 3 or less, it returns the string truncated without ellipsis.
// Otherwise, it truncates the string and appends "..." to indicate truncation.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
