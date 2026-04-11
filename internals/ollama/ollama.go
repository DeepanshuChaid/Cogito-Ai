package ollama

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	sqlite "github.com/DeepanshuChaid/Cogito-Ai-Memo.git/internals/database/sqlite"
)

func distillInteractions() {
	for {
		rows, _ := sqlite.DB.Query("SELECT id, raw_content FROM interactions WHERE processed = 0 LIMIT 10")

		for rows.Next() {
			var id string
			var content string
			rows.Scan(&id, &content)

			fact := askOllamaToSummarize(content)

			sqlite.DB.Exec("INSERT INTO memory (timestamp, fact, importance) VALUES (?, ?, ?)", time.Now(), fact, 5)
			sqlite.DB.Exec("UPDATE interactions SET processed = 1 WHERE id = ?", id)
		}

		rows.Close()
		time.Sleep(10 * time.Second)
	}
}

func askOllamaToSummarize(text string) string {
	payload := map[string]any{
		"model": "llama3.2",
		"prompt": "Summarize the following AI interaction into one single engineering fact. Example: 'Fixed race condition in hub.go using mutex' Text:  " + text,
	}
	 jsonPayload, _ := json.Marshal(payload)

	 response, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(jsonPayload))
	 if err != nil {
		return "Error distilling"
	 }

	 defer response.Body.Close()

	 var result map[string]any
	 json.NewDecoder(response.Body).Decode(&result)
	 return result["response"].(string)
}

func generateMemo() (string, error) {
	rows, err := sqlite.DB.Query("SELECT fact FROM memory ORDER BY timestamp DESC LIMIT 10")
	defer rows.Close()

	if err != nil {
		return "", err
	}

	var facts []string
	for rows.Next() {
		var fact string
		rows.Scan(&fact)
		facts = append(facts, fact)
	}
	return "ENGINEERING MEMORY:\n" + strings.Join(facts, "\n"), nil
}
