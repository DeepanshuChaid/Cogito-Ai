package db

import (
	"database/sql"
	_ "embed"
	"os"
	"path/filepath"
	"time"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/models/schemaModels"
	_ "modernc.org/sqlite" // Pure Go driver (no CGO required)
)

//go:embed schema.sql
var schemaSQL string

var DB *sql.DB

func InitDB() error {
	// 1. Setup the config directory (~/.cogito)
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dirPath := filepath.Join(home, ".cogito")
	dbPath := filepath.Join(dirPath, "cogito.db")

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}

	// 2. Open the database using the "sqlite" driver
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	// 3. Execute the embedded schema
	if _, err := DB.Exec(schemaSQL); err != nil {
		return err
	}

	return nil
}

func GetRelevantObservations(cwd string, limit int) ([]schemaModels.Observations, error) {
	// 1. Use a SQL query that:
	// - Filters by the current project/folder
	// - Prefers CompressedText over RawText
	// - Sorts by newest first
	// - Limits the number of results to avoid token overflow

	query := `
		SELECT id, session_id, project, observation_type, raw_text, compressed_text, files_touched, created_at
		FROM observations
		WHERE project = ?
		ORDER BY created_at DESC
		Limit ?
	`

	rows, err := DB.Query(query, cwd, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var observations []schemaModels.Observations

	for rows.Next() {
		var obs schemaModels.Observations
		var createdAt string
		err := rows.Scan(&obs.ID, &obs.SessionID, &obs.Project, &obs.ObservationType, &obs.RawText, &obs.CompressedText, &obs.FilesTouched, &createdAt)
		if err != nil {
			return nil, err
		}
		obs.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		observations = append(observations, obs)
	}

	return observations, nil
}

// Close shuts down the database connection pool.
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
