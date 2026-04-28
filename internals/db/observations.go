package db

import (
	"time"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/models/schemaModels"
)

// CreateObservation writes durable memory directly to DB
func CreateObservation(sessionID, project, memory, facts string) error {
	now := time.Now()

	_, err := DB.Exec(`
		INSERT INTO observations (
			session_id,
			project,
			memory,
			facts,
			created_at
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		sessionID,
		project,
		memory,
		facts,
		now.Format("2006-01-02 15:04:05"),
	)

	return err
}

// GetRecentObservations fetches latest observations for prompt context
func GetRecentObservations(project string, limit int) ([]schemaModels.Observation, error) {
	rows, err := DB.Query(`
		SELECT
			id,
			session_id,
			project,
			memory,
			facts,
			created_at
		FROM observations
		WHERE project = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, project, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var observations []schemaModels.Observation

	for rows.Next() {
		var o schemaModels.Observation
		var createdAt string

		err := rows.Scan(
			&o.ID,
			&o.SessionID,
			&o.Project,
			&o.Memory,
			&o.Facts,
			&createdAt,
		)

		if err != nil {
			continue
		}

		o.CreatedAt, _ = time.Parse(
			"2006-01-02 15:04:05",
			createdAt,
		)

		observations = append(observations, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return observations, nil
}

// SearchObservationsFTS performs fast keyword lookup using FTS5
func SearchObservationsFTS(project, query string, limit int) ([]schemaModels.ObservationSearchResult, error) {
	rows, err := DB.Query(`
		SELECT
			o.id,
			o.memory,
			bm25(observations_fts) as rank
		FROM observations o
		JOIN observations_fts
			ON o.id = observations_fts.rowid
		WHERE
			o.project = ?
			AND observations_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`, project, query, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []schemaModels.ObservationSearchResult

	for rows.Next() {
		var r schemaModels.ObservationSearchResult

		err := rows.Scan(
			&r.ID,
			&r.Memory,
			&r.Rank,
		)
		if err != nil {
			continue
		}

		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}


// GetObservationByID fetches one full observation
func GetObservationByID(id int) (*schemaModels.Observation, error) {
	row := DB.QueryRow(`
		SELECT
			id,
			session_id,
			project,
			memory,
			facts,
			created_at
		FROM observations
		WHERE id = ?
	`, id)

	var o schemaModels.Observation
	var createdAt string

	err := row.Scan(
		&o.ID,
		&o.SessionID,
		&o.Project,
		&o.Memory,
		&o.Facts,
		&createdAt,
	)

	if err != nil {
		return nil, err
	}

	o.CreatedAt, _ = time.Parse(
		"2006-01-02 15:04:05",
		createdAt,
	)

	return &o, nil
}
