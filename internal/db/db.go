package db

import (
	"database/sql"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/gkwa/sunlitsparrow/internal/logger"
	_ "github.com/mattn/go-sqlite3"
)

// OpenMaccyDB opens a connection to the Maccy database
func OpenMaccyDB() (*sql.DB, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("error getting current user: %w", err)
	}

	// Try multiple possible paths for Maccy database
	possiblePaths := []string{
		filepath.Join(usr.HomeDir, "Library", "Application Support", "Maccy", "Storage.sqlite"),
		filepath.Join(usr.HomeDir, "Library", "Containers", "org.p0deje.Maccy", "Data", "Library", "Application Support", "Maccy", "Storage.sqlite"),
		filepath.Join(usr.HomeDir, "Library", "Group Containers", "43Q936XBMJ.org.p0deje.Maccy", "Library", "Application Support", "Maccy", "Storage.sqlite"),
	}

	// Add the current directory for testing purposes
	possiblePaths = append(possiblePaths, "Maccy-Storage.sqlite")

	// Try each path
	var foundPath string
	for _, path := range possiblePaths {
		logger.Debug("Checking database path: %s", path)
		if _, err := os.Stat(path); err == nil {
			foundPath = path
			logger.Info("Found Maccy database at: %s", foundPath)
			break
		}
	}

	// If no database found, return error
	if foundPath == "" {
		logger.Info("No Maccy database found in any expected location")
		return nil, fmt.Errorf("Maccy database not found in any expected location. You can place a database file named 'Maccy-Storage.sqlite' in the current directory for testing")
	}

	db, err := sql.Open("sqlite3", foundPath)
	if err != nil {
		return nil, fmt.Errorf("error opening SQLite database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	logger.Info("Successfully connected to Maccy database")
	return db, nil
}
