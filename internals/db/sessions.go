package db

import (
	"database/sql"
	"time"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/models/schemaModels"
)

// CreateSession creates a new session or returns existing one (idempotent)
// Called by: context-hook (SessionStart)
func CreateSession(contentSessionID, project, userPrompt string) (*schemaModels.Session, error) {
	now :=  time.Now()

	existing := &schemaModels.Session{}
	err := DB.QueryRow(`
		SELECT id, content_session_id, memory_session_id, project, status, user_prompt, started_at, completed_at
		FROM sdk_sessions
		WHERE content_session_id = ?
	`, contentSessionID).Scan(
		&existing.ID, &existing.ContentSessionID, &existing.MemorySessionID,
		&existing.Project, &existing.Status, &existing.UserPrompt,
		&existing.StartedAt, &existing.CompletedAt,
	)

	if err == nil {
		// SESSION EXISTS, RETURN IT
		return existing, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	result, err := DB.Exec(`
		INSERT INTO sdk_sessions (content_session_id, memory_session_id, project, status, user_prompt, started_at)
		VALUES (?, NULL, ?, 'active', ?, ?)
	`, contentSessionID, project, userPrompt, now.Format("2006-01-02 15:04:05"))

	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()

	return &schemaModels.Session{
		ID: int(id),
		ContentSessionID: contentSessionID,
		Project: project,
		Status: "active",
		UserPrompt: userPrompt,
		StartedAt: now,
	}, nil
}

// UpdateMemorySessionID links a Cogito memory_session_id to the session
// Called by: Worker after first SDK response
func UpdateMemorySessionID(sessionID int, memorySessionID string) error {
	_, err := DB.Exec(`
		UPDATE sdk_sessions
		SET memory_session_id = ?
		WHERE id = ?
	`, memorySessionID, sessionID)

	return err
}

// CompleteSession marks a session as completed
// Called by: summary-hook (SessionEnd)
func CompleteSession(contentSessionID string) error {
	now := time.Now()
	_, err := DB.Exec(`
		UPDATE sdk_sessions
		SET status = 'completed', completed_at = ?
		WHERE content_session_id = ?
	`, now.Format("2006-01-02 15:04:05"), contentSessionID)
	return err
}

// GetSessionByContentID retrieves a session by IDE/Codex session ID
func GetSessionByContentID(contentSessionID string) (*schemaModels.Session, error) {
	session := &schemaModels.Session{}
	err := DB.QueryRow(`
		SELECT id, content_session_id, memory_session_id, project, status, user_prompt, started_at, completed_at
		FROM sdk_sessions
		WHERE content_session_id = ?
	`, contentSessionID).Scan(
		&session.ID, &session.ContentSessionID, &session.MemorySessionID,
		&session.Project, &session.Status, &session.UserPrompt,
		&session.StartedAt, &session.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return session, err
}
