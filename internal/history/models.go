package history

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"time"
)

// HistoryItem represents a clipboard history item
type HistoryItem struct {
	ID             int       `json:"id"`
	Title          string    `json:"title"`
	Pin            string    `json:"pin,omitempty"`
	FirstCopiedAt  time.Time `json:"firstCopiedAt"`
	LastCopiedAt   time.Time `json:"lastCopiedAt"`
	NumberOfCopies int       `json:"numberOfCopies"`
	Application    string    `json:"application,omitempty"`
	Contents       []Content `json:"contents,omitempty"`
}

// Content represents the content of a history item
type Content struct {
	Type  string `json:"type"`
	Value []byte `json:"-"` // Don't include raw bytes in JSON
}

// MarshalJSON customizes JSON marshaling for Content
func (c Content) MarshalJSON() ([]byte, error) {
	// For JSON, we'll encode binary data as base64
	type ContentAlias struct {
		Type  string `json:"type"`
		Value string `json:"value"` // Base64 encoded value
	}

	// Special handling for different content types
	var valueStr string
	switch c.Type {
	case "public.utf8-plain-text":
		// For text content, convert directly to string
		valueStr = string(c.Value)
	default:
		// For binary content, use base64 encoding
		valueStr = base64.StdEncoding.EncodeToString(c.Value)
	}

	return json.Marshal(ContentAlias{
		Type:  c.Type,
		Value: valueStr,
	})
}

// NullableHistoryItem is used for scanning SQL results with potential NULL values
type NullableHistoryItem struct {
	ID             int
	Title          sql.NullString
	Pin            sql.NullString
	FirstCopiedAt  sql.NullFloat64 // Changed from sql.NullTime to handle float timestamp
	LastCopiedAt   sql.NullFloat64 // Changed from sql.NullTime to handle float timestamp
	NumberOfCopies sql.NullInt64
	Application    sql.NullString
}

// ToHistoryItem converts a NullableHistoryItem to a HistoryItem
func (n *NullableHistoryItem) ToHistoryItem() HistoryItem {
	item := HistoryItem{
		ID: n.ID,
	}

	if n.Title.Valid {
		item.Title = n.Title.String
	}
	if n.Pin.Valid {
		item.Pin = n.Pin.String
	}
	if n.FirstCopiedAt.Valid {
		// Convert Cocoa Core Data timestamp (seconds since Jan 1, 2001) to Go time
		item.FirstCopiedAt = cocoaTimestampToTime(n.FirstCopiedAt.Float64)
	}
	if n.LastCopiedAt.Valid {
		// Convert Cocoa Core Data timestamp (seconds since Jan 1, 2001) to Go time
		item.LastCopiedAt = cocoaTimestampToTime(n.LastCopiedAt.Float64)
	}
	if n.NumberOfCopies.Valid {
		item.NumberOfCopies = int(n.NumberOfCopies.Int64)
	}
	if n.Application.Valid {
		item.Application = n.Application.String
	}

	return item
}

// cocoaTimestampToTime converts a Cocoa/Core Data timestamp (seconds since Jan 1, 2001)
// to a Go time.Time
func cocoaTimestampToTime(timestamp float64) time.Time {
	// Cocoa reference date is January 1, 2001, 00:00:00 UTC
	referenceDate := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)

	// Convert seconds to duration
	duration := time.Duration(timestamp * float64(time.Second))

	// Add duration to reference date
	return referenceDate.Add(duration)
}
