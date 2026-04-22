package db

import (
	"database/sql"
	"time"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/models/schemaModels"
)

// InitializeProjectSession handles the "Handshake" from the AI tool.
func InitializeProjectSession(toolSessionID, absPath string) (*schemaModels.Session, error) {
	now := time.Now()

	existing := &schemaModels.Session{}
	var startedAtStr string

	err := DB.QueryRow(`
		SELECT id, session_id, project, started_at
		FROM sdk_sessions
		WHERE session_id = ?
	`, toolSessionID).Scan(
		&existing.ID,
		&existing.SessionID,
		&existing.Project,
		&startedAtStr,
	)

	if err == nil {
		// Convert string → time.Time
		parsedTime, parseErr := time.Parse("2006-01-02 15:04:05", startedAtStr)
		if parseErr == nil {
			existing.StartedAt = parsedTime
		}
		return existing, nil
	}

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// ✅ FIXED INSERT (you had a bug here too)
	res, err := DB.Exec(`
		INSERT INTO sdk_sessions (session_id, project, started_at)
		VALUES (?, ?, ?)
	`, toolSessionID, absPath, now)

	if err != nil {
		return nil, err
	}

	id, _ := res.LastInsertId()

	return &schemaModels.Session{
		ID:        int(id),
		SessionID: toolSessionID,
		Project:   absPath,
		StartedAt: now,
	}, nil
}


// UpdateMemorySessionID links a Cogito memory_session_id to the session
func UpdateMemorySessionID(sessionID int, memorySessionID string) error {
	_, err := DB.Exec(`
		UPDATE sdk_sessions
		SET memory_session_id = ?
		WHERE id = ?
	`, memorySessionID, sessionID)

	return err
}

// CompleteSession marks a session as completed
func CompleteSession(sessionID string) error {
	now := time.Now()

	_, err := DB.Exec(`
		UPDATE sdk_sessions
		SET completed_at = ?
		WHERE session_id = ?
	`, now, sessionID)

	return err
}
