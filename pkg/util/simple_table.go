package util

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// DisplaySimpleTable creates a clean, borderless table with better formatting
func DisplaySimpleTable(data []map[string]interface{}, headers []string) {
	if len(data) == 0 {
		fmt.Println("No data to display.")
		return
	}

	// Calculate column widths
	colWidths := make([]int, len(headers))

	// Initialize with header widths
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	// Check data widths
	for _, row := range data {
		for i, header := range headers {
			if val, ok := row[header]; ok {
				var valueStr string
				switch v := val.(type) {
				case string:
					valueStr = v
				case int:
					valueStr = fmt.Sprintf("%d", v)
				case float64:
					valueStr = fmt.Sprintf("%.2f", v)
				default:
					valueStr = fmt.Sprintf("%v", v)
				}

				// Limit max column width for readability
				maxWidth := 60
				if len(valueStr) > maxWidth {
					valueStr = valueStr[:maxWidth-3] + "..."
				}

				if len(valueStr) > colWidths[i] {
					colWidths[i] = len(valueStr)
				}
			}
		}
	}

	// Add some spacing
	fmt.Println()

	// Print headers with color but keep original case
	headerColor := color.New(color.FgBlue, color.Bold)
	for i, header := range headers {
		if i > 0 {
			fmt.Print("  ")
		}
		// Print colored header and then pad with spaces to reach the column width
		coloredHeader := headerColor.Sprint(header)
		fmt.Print(coloredHeader)
		// Add padding to match column width (account for the actual header length, not the colored string length)
		padding := colWidths[i] - len(header)
		if padding > 0 {
			fmt.Print(strings.Repeat(" ", padding))
		}
	}
	fmt.Println()

	// Print separator line
	for i := range headers {
		if i > 0 {
			fmt.Print("  ")
		}
		fmt.Print(strings.Repeat("─", colWidths[i]))
	}
	fmt.Println()

	// Print data rows
	gray := color.New(color.FgHiBlack)
	for _, row := range data {
		for i, header := range headers {
			if i > 0 {
				fmt.Print("  ")
			}

			var valueStr string
			if val, ok := row[header]; ok {
				switch v := val.(type) {
				case string:
					valueStr = v
				case int:
					valueStr = fmt.Sprintf("%d", v)
				case float64:
					valueStr = fmt.Sprintf("%.2f", v)
				default:
					valueStr = fmt.Sprintf("%v", v)
				}
			} else {
				valueStr = gray.Sprint("-")
			}

			// Truncate if too long
			maxWidth := 60
			if len(valueStr) > maxWidth {
				valueStr = valueStr[:maxWidth-3] + "..."
			}

			fmt.Printf("%-*s", colWidths[i], valueStr)
		}
		fmt.Println()
	}

	// Add spacing after
	fmt.Println()
}
