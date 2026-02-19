package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func Success(data interface{}) {
	out := map[string]interface{}{
		"ok":   true,
		"data": data,
	}
	printJSON(out)
}

func SuccessList(data interface{}, count int) {
	out := map[string]interface{}{
		"ok":    true,
		"count": count,
		"data":  data,
	}
	printJSON(out)
}

func Error(msg string) {
	out := map[string]interface{}{
		"ok":    false,
		"error": msg,
	}
	printJSON(out)
}

func ErrorDetail(msg string, detail interface{}) {
	out := map[string]interface{}{
		"ok":     false,
		"error":  msg,
		"detail": detail,
	}
	printJSON(out)
}

func Raw(data interface{}) {
	printJSON(data)
}

func printJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

// Table output

func Table(headers []string, rows [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetBorder(true)
	table.SetRowLine(false)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.AppendBulk(rows)
	table.Render()
}

// CSV output

func CSV(headers []string, rows [][]string) {
	w := csv.NewWriter(os.Stdout)
	_ = w.Write(headers)
	for _, row := range rows {
		_ = w.Write(row)
	}
	w.Flush()
}

// FormatTable is a helper that maps generic data to table format
func FormatTable(format string, headers []string, rows [][]string) {
	switch strings.ToLower(format) {
	case "table":
		Table(headers, rows)
	case "csv":
		CSV(headers, rows)
	default:
		// JSON is handled by caller
	}
}
