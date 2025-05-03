package history

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gkwa/sunlitsparrow/internal/logger"
)

// Repository handles database operations for history items
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new history repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetRecentItems retrieves the most recent history items
func (r *Repository) GetRecentItems(limit int) ([]HistoryItem, error) {
	// Try different schema versions
	items, err := r.tryStandardSchema(limit)
	if err == nil {
		return items, nil
	}

	logger.Debug("Standard schema query failed, trying alternative: %v", err)
	items, err = r.tryAlternativeSchema(limit)
	if err == nil {
		return items, nil
	}

	logger.Debug("Alternative schema query failed, trying dynamic approach: %v", err)
	return r.tryDynamicSchema(limit)
}

// GetAllItems retrieves all history items
func (r *Repository) GetAllItems() ([]HistoryItem, error) {
	// Similar to GetRecentItems but without limit
	items, err := r.tryStandardSchema(0)
	if err == nil {
		return items, nil
	}

	logger.Debug("Standard schema query failed, trying alternative: %v", err)
	items, err = r.tryAlternativeSchema(0)
	if err == nil {
		return items, nil
	}

	logger.Debug("Alternative schema query failed, trying dynamic approach: %v", err)
	return r.tryDynamicSchema(0)
}

// GetPinnedItems retrieves all pinned history items
func (r *Repository) GetPinnedItems() ([]HistoryItem, error) {
	// Try standard schema query for pinned items
	query := `
		SELECT id, title, pin, firstCopiedAt, lastCopiedAt, numberOfCopies, application
		FROM HistoryItem
		WHERE pin IS NOT NULL AND pin != ''
		ORDER BY lastCopiedAt DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		// Try alternative schema query
		logger.Debug("Standard schema query for pins failed, trying alternative: %v", err)
		query = `
			SELECT Z_PK, ZTITLE, ZPIN, ZFIRSTCOPIEDAT, ZLASTCOPIEDAT, ZNUMBEROFCOPIES, ZAPPLICATION
			FROM ZHISTORYITEM
			WHERE ZPIN IS NOT NULL AND ZPIN != ''
			ORDER BY ZLASTCOPIEDAT DESC
		`
		rows, err = r.db.Query(query)
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()

	return r.scanHistoryItems(rows)
}

// GetItemContents retrieves contents for a specific history item
func (r *Repository) GetItemContents(itemID int) ([]Content, error) {
	// Try standard schema
	contentRows, err := r.db.Query(`
		SELECT type, value
		FROM HistoryItemContent
		WHERE item_id = ?
	`, itemID)
	if err != nil {
		// Try alternative schema
		contentRows, err = r.db.Query(`
			SELECT ZTYPE, ZVALUE
			FROM ZHISTORYITEMCONTENT
			WHERE ZITEM = ?
		`, itemID)
		if err != nil {
			return nil, err
		}
	}
	defer contentRows.Close()

	var contents []Content

	for contentRows.Next() {
		var content Content
		var contentType string
		var contentValue []byte

		if err := contentRows.Scan(&contentType, &contentValue); err != nil {
			return nil, err
		}

		content.Type = contentType
		content.Value = contentValue
		contents = append(contents, content)
	}

	return contents, nil
}

func (r *Repository) tryStandardSchema(limit int) ([]HistoryItem, error) {
	query := `
		SELECT id, title, pin, firstCopiedAt, lastCopiedAt, numberOfCopies, application
		FROM HistoryItem
		ORDER BY lastCopiedAt DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanHistoryItems(rows)
}

func (r *Repository) tryAlternativeSchema(limit int) ([]HistoryItem, error) {
	query := `
		SELECT Z_PK, ZTITLE, ZPIN, ZFIRSTCOPIEDAT, ZLASTCOPIEDAT, ZNUMBEROFCOPIES, ZAPPLICATION
		FROM ZHISTORYITEM
		ORDER BY ZLASTCOPIEDAT DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanHistoryItems(rows)
}

func (r *Repository) tryDynamicSchema(limit int) ([]HistoryItem, error) {
	// Check if HistoryItem table exists
	var tableExists bool
	err := r.db.QueryRow(`
		SELECT COUNT(*) > 0 FROM sqlite_master
		WHERE type='table' AND name='HistoryItem'
	`).Scan(&tableExists)
	if err != nil {
		return nil, fmt.Errorf("error checking for HistoryItem table: %w", err)
	}

	if !tableExists {
		return nil, fmt.Errorf("HistoryItem table not found in database")
	}

	// Get column names
	rows, err := r.db.Query(`SELECT * FROM HistoryItem LIMIT 1`)
	if err != nil {
		return nil, err
	}
	cols, err := rows.Columns()
	rows.Close()
	if err != nil {
		return nil, err
	}

	logger.Debug("HistoryItem columns: %s", strings.Join(cols, ", "))

	// Build a dynamic query based on actual column names
	columnsNeeded := map[string]string{
		"id":             "id",
		"title":          "title",
		"pin":            "pin",
		"firstCopiedAt":  "firstCopiedAt",
		"lastCopiedAt":   "lastCopiedAt",
		"numberOfCopies": "numberOfCopies",
		"application":    "application",
	}

	selectCols := make([]string, 0, len(columnsNeeded))
	for _, col := range cols {
		for standardName := range columnsNeeded {
			if strings.EqualFold(col, standardName) || strings.EqualFold(col, "Z"+standardName) {
				selectCols = append(selectCols, col)
				break
			}
		}
	}

	if len(selectCols) == 0 {
		return nil, fmt.Errorf("couldn't identify required columns in HistoryItem table")
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM HistoryItem
		ORDER BY lastCopiedAt DESC, firstCopiedAt DESC
	`, strings.Join(selectCols, ", "))

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err = r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanHistoryItems(rows)
}

func (r *Repository) scanHistoryItems(rows *sql.Rows) ([]HistoryItem, error) {
	var items []HistoryItem

	for rows.Next() {
		var nullableItem NullableHistoryItem

		if err := rows.Scan(
			&nullableItem.ID,
			&nullableItem.Title,
			&nullableItem.Pin,
			&nullableItem.FirstCopiedAt,
			&nullableItem.LastCopiedAt,
			&nullableItem.NumberOfCopies,
			&nullableItem.Application,
		); err != nil {
			logger.Debug("Error scanning row: %v", err)
			continue
		}

		item := nullableItem.ToHistoryItem()

		// Get contents for this item
		contents, err := r.GetItemContents(item.ID)
		if err != nil {
			logger.Debug("Error getting contents for item %d: %v", item.ID, err)
		} else {
			item.Contents = contents
		}

		items = append(items, item)
	}

	return items, nil
}
