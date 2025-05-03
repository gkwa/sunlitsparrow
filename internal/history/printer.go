package history

import (
	"fmt"
	"strings"
	"time"
)

// Printer handles printing of history items
type Printer struct {
	items []HistoryItem
}

// NewPrinter creates a new history item printer
func NewPrinter(items []HistoryItem) *Printer {
	return &Printer{items: items}
}

// PrintItems prints history items in a formatted table
func (p *Printer) PrintItems() {
	if len(p.items) == 0 {
		fmt.Println("No items found.")
		return
	}

	fmt.Printf("%-5s %-20s %-3s %-30s %-30s %-5s %-20s\n",
		"ID", "Title", "Pin", "First Copied", "Last Copied", "Count", "Application")
	fmt.Println(strings.Repeat("-", 120))

	for _, item := range p.items {
		titleStr := item.Title
		if len(titleStr) > 20 {
			titleStr = titleStr[:17] + "..."
		}

		pinStr := item.Pin
		if pinStr == "" {
			pinStr = "-"
		}

		appStr := item.Application
		if appStr == "" {
			appStr = "<unknown>"
		}

		firstCopiedStr := formatTime(item.FirstCopiedAt)
		lastCopiedStr := formatTime(item.LastCopiedAt)

		fmt.Printf("%-5d %-20s %-3s %-30s %-30s %-5d %-20s\n",
			item.ID, titleStr, pinStr, firstCopiedStr, lastCopiedStr, item.NumberOfCopies, appStr)
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "<never>"
	}
	// For timestamp zero values from Cocoa (1970-01-01)
	if t.Year() == 1970 && t.Month() == 1 && t.Day() == 1 {
		return "<never>"
	}
	// Prefer local time for display
	return t.Local().Format("2006-01-02 15:04:05")
}
