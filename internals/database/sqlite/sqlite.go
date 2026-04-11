package sqlite

import (
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	DB, err := sql.Open("sqlite", "./cogito.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS interaction (
			id UUID PRIMARY KEY,
			timestamt DATETIME,
			raw_content TEXT,
			processed INTEGER DEFAULT 0 -- 0 needs distillation, 1 = sumarized
		)
	`)
	if err != nil {
		log.Fatal("INTERACTION TABLE", err)
	}

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS memory (
			id UUID PRIMARY KEY,
			timestamp DATETIME,
			fact TEXT,
			importance INTEGER -- 1 to 10
		)
	`)
	if err != nil {
		log.Fatal("memory table", err)
	}



	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS requests (
			id UUID PRIMARY KEY,
			timestamp DATETIME,
			method TEXT,
			path TEXT,
			body_preview TEXT,
			tokens INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database Initialized")
}

func LogRequest(method, path, body string) {
	// TRUNCATE BODY FOR PREVIEW (FIRST 200 CHARS)
	preview := body
	if len(body) > 200 {
		preview = body[:200] + "..."
	}

	_, err := DB.Exec(
		"INSERT INTO requests (timestamp, method, path, body_preview) VALUES (?, ?, ?, ?)",
		time.Now(),
		method, path, preview,
	)
	if err != nil {
		log.Println("Failed to log: ", err)
	}
}

