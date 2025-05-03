package export

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gkwa/sunlitsparrow/internal/history"
)

// JSONExporter handles exporting of history items to JSON
type JSONExporter struct {
	outputFile string
}

// NewJSONExporter creates a new JSON exporter
func NewJSONExporter(outputFile string) *JSONExporter {
	return &JSONExporter{outputFile: outputFile}
}

// Export exports history items to a JSON file
func (e *JSONExporter) Export(items []history.HistoryItem) error {
	file, err := os.Create(e.outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(items); err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}

	return nil
}
