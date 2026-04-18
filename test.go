package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// Test prints the contents of the three core tables using the
// schema you posted (sdk_sessions, observations, session_summaries).
func main() {
	// -----------------------------------------------------------------
	// 1️⃣ Resolve the exact DB path (same as db.InitDB)
	// -----------------------------------------------------------------
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".cogito", "cogito.db")

	fmt.Printf("🔍 Inspecting Database: %s\n", dbPath)
	fmt.Println(strings.Repeat("=", 80))

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("❌ Could not open database: %v", err)
	}
	defer db.Close()

	// -----------------------------------------------------------------
	// 2️⃣ Print each table
	// -----------------------------------------------------------------
	printSessions(db)
	printObservations(db)
	printSummaries(db)
}

// ---------------------------------------------------------------------
// 2️⃣ Sessions (sdk_sessions)
// ---------------------------------------------------------------------
func printSessions(db *sql.DB) {
	fmt.Println("\n📂 [SDK_SESSIONS]")
	rows, err := db.Query(`
		SELECT id,
		       content_session_id,
		       project,
		       status,
		       user_prompt,
		       started_at,
		       COALESCE(completed_at, '')
		FROM sdk_sessions
		ORDER BY started_at DESC
	`)
	if err != nil {
		fmt.Printf("⚠️  Error fetching sessions: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Printf("%-4s | %-20s | %-30s | %-10s | %-30s | %-20s | %-20s\n",
		"ID", "ContentSID", "Project", "Status", "Prompt", "StartedAt", "CompletedAt")
	fmt.Println(strings.Repeat("-", 120))

	for rows.Next() {
		var (
			id               int
			contentSID, proj string
			status, prompt   string
			started, comp    string
		)

		if err := rows.Scan(&id, &contentSID, &proj, &status, &prompt, &started, &comp); err != nil {
			fmt.Printf("⚠️  Scan error: %v\n", err)
			continue
		}
		fmt.Printf("%-4d | %-20s | %-30s | %-10s | %-30.30s | %-20s | %-20s\n",
			id, contentSID, proj, status, prompt, started, comp)
	}
}

// ---------------------------------------------------------------------
// 3️⃣ Observations (observations)
// ---------------------------------------------------------------------
func printObservations(db *sql.DB) {
	fmt.Println("\n🧠 [OBSERVATIONS]")
	rows, err := db.Query(`
		SELECT id,
		       obs_type,
		       project,
		       title,
		       compressed_text,
		       facts,
		       files_touched,
		       discovery_tokens,
		       created_at
		FROM observations
		ORDER BY created_at DESC
	`)
	if err != nil {
		fmt.Printf("⚠️  Error fetching observations: %v\n", err)
		return
	}
	defer rows.Close()

	header := fmt.Sprintf("%-4s | %-10s | %-30s | %-30s | %-10s | %-20s | %-20s | %-20s",
		"ID", "Type", "Project", "Title", "Tokens", "CreatedAt", "Facts", "Files")
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", 150))

	for rows.Next() {
		var (
			id               int
			obsType, proj    string
			title, comp      string
			facts, files     string
			tokens           int
			createdAt        string
		)

		if err := rows.Scan(&id, &obsType, &proj, &title, &comp, &facts, &files, &tokens, &createdAt); err != nil {
			fmt.Printf("⚠️  Scan error: %v\n", err)
			continue
		}

		// Trim long strings for console readability
		if len(title) > 25 {
			title = title[:22] + "..."
		}
		if len(comp) > 25 {
			comp = comp[:22] + "..."
		}
		if len(facts) > 25 {
			facts = facts[:22] + "..."
		}
		if len(files) > 25 {
			files = files[:22] + "..."
		}

		fmt.Printf("%-4d | %-10s | %-30s | %-30s | %-10d | %-20s | %-20s | %-20s\n",
			id, obsType, proj, title, tokens, createdAt, facts, files)
	}
}

// ---------------------------------------------------------------------
// 4️⃣ Summaries (session_summaries)
// ---------------------------------------------------------------------
func printSummaries(db *sql.DB) {
	fmt.Println("\n📝 [SESSION_SUMMARIES]")
	rows, err := db.Query(`
		SELECT id,
		       memory_session_id,
		       project,
		       request,
		       learned,
		       next_steps,
		       created_at
		FROM session_summaries
		ORDER BY created_at DESC
	`)
	if err != nil {
		fmt.Printf("⚠️  Error fetching summaries: %v\n", err)
		return
	}
	defer rows.Close()

	header := fmt.Sprintf("%-4s | %-20s | %-30s | %-30s | %-30s | %-30s | %-20s",
		"ID", "MemSID", "Project", "Request", "Learned", "NextSteps", "CreatedAt")
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", 150))

	for rows.Next() {
		var (
			id                              int
			memSID, proj, request, learned   string
			nextSteps, createdAt             string
		)

		if err := rows.Scan(&id, &memSID, &proj, &request, &learned, &nextSteps, &createdAt); err != nil {
			fmt.Printf("⚠️  Scan error: %v\n", err)
			continue
		}

		// Trim very long text for display
		if len(request) > 25 {
			request = request[:22] + "..."
		}
		if len(learned) > 25 {
			learned = learned[:22] + "..."
		}
		if len(nextSteps) > 25 {
			nextSteps = nextSteps[:22] + "..."
		}
		fmt.Printf("%-4d | %-20s | %-30s | %-30s | %-30s | %-30s | %-20s\n",
			id, memSID, proj, request, learned, nextSteps, createdAt)
	}
}
