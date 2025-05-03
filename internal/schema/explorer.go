package schema

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/gkwa/sunlitsparrow/internal/logger"
)

// Explorer handles schema exploration of the database
type Explorer struct {
	db *sql.DB
}

// NewExplorer creates a new schema explorer
func NewExplorer(db *sql.DB) *Explorer {
	return &Explorer{db: db}
}

// ExploreSchema prints the schema of the Maccy database
func (e *Explorer) ExploreSchema() {
	// Get list of tables
	rows, err := e.db.Query(`
		SELECT name FROM sqlite_master
		WHERE type='table' AND name NOT LIKE 'sqlite_%'
	`)
	if err != nil {
		fmt.Printf("Error querying tables: %v\n", err)
		return
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			logger.Debug("Error scanning table name: %v", err)
			continue
		}
		tables = append(tables, tableName)
	}

	// Print table schemas
	for _, table := range tables {
		logger.Info("Table: %s", table)

		// Get table schema
		e.printTableSchema(table)
		e.printForeignKeys(table)
		e.printIndices(table)

		fmt.Println() // Empty line between tables
	}
}

// ExportSchemaToFile exports the schema to a SQLite-compatible file
func (e *Explorer) ExportSchemaToFile(filename string) error {
	// Get list of tables
	rows, err := e.db.Query(`
		SELECT sql FROM sqlite_master
		WHERE type='table' AND name NOT LIKE 'sqlite_%'
	`)
	if err != nil {
		return fmt.Errorf("error querying tables: %w", err)
	}
	defer rows.Close()

	// Collect all CREATE TABLE statements
	var createTableStatements []string
	for rows.Next() {
		var sql string
		if err := rows.Scan(&sql); err != nil {
			logger.Debug("Error scanning SQL: %v", err)
			continue
		}
		createTableStatements = append(createTableStatements, sql+";")
	}

	// Get list of indices
	indexRows, err := e.db.Query(`
		SELECT sql FROM sqlite_master
		WHERE type='index' AND name NOT LIKE 'sqlite_%' AND sql IS NOT NULL
	`)
	if err != nil {
		return fmt.Errorf("error querying indices: %w", err)
	}
	defer indexRows.Close()

	// Collect all CREATE INDEX statements
	var createIndexStatements []string
	for indexRows.Next() {
		var sql string
		if err := indexRows.Scan(&sql); err != nil {
			logger.Debug("Error scanning index SQL: %v", err)
			continue
		}
		createIndexStatements = append(createIndexStatements, sql+";")
	}

	// Write to file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer file.Close()

	// Add transaction wrapper
	file.WriteString("BEGIN TRANSACTION;\n\n")

	// Write all CREATE TABLE statements
	for _, stmt := range createTableStatements {
		file.WriteString(stmt + "\n\n")
	}

	// Write all CREATE INDEX statements
	for _, stmt := range createIndexStatements {
		file.WriteString(stmt + "\n\n")
	}

	file.WriteString("COMMIT;\n")

	return nil
}

// printTableSchema prints the schema of a table
func (e *Explorer) printTableSchema(table string) {
	pragma, err := e.db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		logger.Debug("Error getting schema for table %s: %v", table, err)
		return
	}
	defer pragma.Close()

	for pragma.Next() {
		var cid int
		var name, columnType string
		var notnull, pk int
		var dfltValue interface{}

		if err := pragma.Scan(&cid, &name, &columnType, &notnull, &dfltValue, &pk); err != nil {
			logger.Trace("Error scanning column data: %v", err)
			continue
		}

		columnDesc := fmt.Sprintf("  - %s (%s)", name, columnType)
		if notnull == 1 {
			columnDesc += " NOT NULL"
		}
		if pk == 1 {
			columnDesc += " PRIMARY KEY"
		}
		if dfltValue != nil {
			columnDesc += fmt.Sprintf(" DEFAULT %v", dfltValue)
		}

		fmt.Println(columnDesc)
	}
}

// printForeignKeys prints the foreign keys of a table
func (e *Explorer) printForeignKeys(table string) {
	fkRows, err := e.db.Query(fmt.Sprintf("PRAGMA foreign_key_list(%s)", table))
	if err != nil {
		logger.Debug("Error getting foreign keys for table %s: %v", table, err)
		return
	}
	defer fkRows.Close()

	for fkRows.Next() {
		var id, seq int
		var toTable, from, to string
		var onUpdate, onDelete string
		var match interface{}

		if err := fkRows.Scan(&id, &seq, &toTable, &from, &to, &onUpdate, &onDelete, &match); err != nil {
			logger.Trace("Error scanning foreign key data: %v", err)
			continue
		}

		fkDesc := fmt.Sprintf("  - Foreign Key: %s -> %s.%s", from, toTable, to)
		fmt.Println(fkDesc)
	}
}

// printIndices prints the indices of a table
func (e *Explorer) printIndices(table string) {
	idxRows, err := e.db.Query(fmt.Sprintf("PRAGMA index_list(%s)", table))
	if err != nil {
		logger.Debug("Error getting indices for table %s: %v", table, err)
		return
	}
	defer idxRows.Close()

	for idxRows.Next() {
		var seqno int
		var name string
		var unique bool
		var origin, partial string

		if err := idxRows.Scan(&seqno, &name, &unique, &origin, &partial); err != nil {
			logger.Trace("Error scanning index data: %v", err)
			continue
		}

		idxDesc := fmt.Sprintf("  - Index: %s", name)
		if unique {
			idxDesc += " (UNIQUE)"
		}
		fmt.Println(idxDesc)
	}
}
