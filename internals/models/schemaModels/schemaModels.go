package schemaModels

import "time"


type Session struct {
	ID             int    `json:"id"`
	SesssionID     string `json:"session_id"`
	Project        string `json:"project"`
	StartedAt      time.Time `json:"started_at"`
	Status         string `json:"status"`
	UserPrompt     string `json:"user_prompt"`
}


type Observations struct {
	ID              int    `json:"id"`
	SessionID       string `json:"session_id"`
	Project         string `json:"project"`
	ObservationType string `json:"observation_type"`
	RawText         string `json:"raw_text"`
	CompressedText  string `json:"compressed_text"` // THIS IS WHAT WILL BE SENT TO AI
	FilesTouched    string `json:"file_touched"` // JSON STRING
	CreatedAt       time.Time `json:"created_at"`
}

type Summaries struct {
	ID 				int `json:"id"`
	SessionID       string `json:"session_id"`
	Project         string `json:"project"`
	SummaryText string `json:"summary_text"`
	LearnedFacts string `json:"learned_facts"`
	CreatedAt time.Time `json:"created_at"`
}
