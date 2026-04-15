package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
)

type Memory struct {
	FilePath       string
	CompressedText string
}

func GetDB() *sql.DB {
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".cogito", "cogito.db")

	os.MkdirAll(filepath.Dir(dbPath), 0755)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS memories (
		file_path TEXT PRIMARY KEY,
		compressed_text TEXT,
		last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(sqlStmt)
	if err != nil{
		log.Fatal(err)
	}

	return db
}

func SaveMemory (filePath, text string) error {
	db := GetDB()
	defer db.Close()

	_, err := db.Exec("INSERT OR REPLACE INTO memories (file_path, compressed_text) VALUES (?, ?)", filePath, text)

	return err
}

func GetAllMemories() ([]Memory, error) {
	db := GetDB()
	defer db.Close()
	rows, err := db.Query("SELECT file_path, compressed_text FROM memories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []Memory
	for rows.Next() {
		var m Memory
		rows.Scan(&m.FilePath, &m.CompressedText)
		memories = append(memories, m)
	}
	return memories, nil
}
