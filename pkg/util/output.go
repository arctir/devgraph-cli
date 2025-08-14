package util

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
)

// DisplayTable takes a slice of maps (data) and headers, and displays it as a formatted table
func DisplayTable(data []map[string]interface{}, headers []string) {
	// Create a new table writer
	table := tablewriter.NewWriter(os.Stdout)

	// Set the headers
	table.Header(headers)

	// Convert map data to table rows
	for _, row := range data {
		var rowData []string
		for _, header := range headers {
			// Convert interface{} to string
			var value string
			if val, ok := row[header]; ok {
				switch v := val.(type) {
				case string:
					value = v
				case int:
					value = fmt.Sprintf("%d", v)
				case float64:
					value = fmt.Sprintf("%.2f", v)
				default:
					value = fmt.Sprintf("%v", v)
				}
			} else {
				value = ""
			}
			rowData = append(rowData, value)
		}
		table.Append(rowData)
	}

	// Set table formatting
	/*table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("|")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("-")
	*/
	// Render the table
	table.Render()
}
