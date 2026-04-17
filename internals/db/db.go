package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type Memory struct {
	FilePath       string
	CompressedText string
}

var DB *sql.DB

func InitDB() error {
	// 1. Determine where to store the DB (e.g., ~/.cogito/cogito.db)
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".cogito", "cogito.db")

	// Ensure the directory exists
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	// 2. Open the database
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	// 3. Read the schema.sql file
	schemaFile := filepath.Join("internals", "db", "schema.sql")
	schema, err := os.ReadFile(schemaFile)
	if err != nil {
		return err
	}

	// 4. Execute the schema to create tables
	_, err = DB.Exec(string(schema))
	if err != nil {
		return err
	}

	return nil
}

