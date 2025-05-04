package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Initialize initializes the SQLite database connection.
func Initialize(dataSourceName string) {
	var err error
	db, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
}

// GetDB returns the database instance.
func GetDB() *sql.DB {
	return db
}
