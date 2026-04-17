package commands

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // Use the same driver as your main app
)

func Test() {
	// 1. Locate the DB file (Must match your InitDB path)
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".cogito", "cogito.db")

	fmt.Printf("🔍 Inspecting Database: %s\n", dbPath)
	fmt.Println("================================================================")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal("❌ Could not open database: ", err)
	}
	defer db.Close()

	// Run all tests
	printSessions(db)
	printObservations(db)
	printSummaries(db)
}

func printSessions(db *sql.DB) {
	fmt.Println("\n📂 [SESSIONS]")
	rows, err := db.Query("SELECT id, session_id, project, status FROM sessions")
	if err != nil {
		fmt.Printf("Error fetching sessions: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Printf("%-4s | %-20s | %-20s | %-10s\n", "ID", "SessionID", "Project", "Status")
	fmt.Println("----------------------------------------------------------------")
	for rows.Next() {
		var id int
		var sid, proj, status string
		rows.Scan(&id, &sid, &proj, &status)
		fmt.Printf("%-4d | %-20s | %-20s | %-10s\n", id, sid, proj, status)
	}
}

func printObservations(db *sql.DB) {
	fmt.Println("\n🧠 [OBSERVATIONS]")
	rows, err := db.Query("SELECT id, obs_type, project, raw_text, compressed_text FROM observations")
	if err != nil {
		fmt.Printf("Error fetching observations: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Printf("%-4s | %-12s | %-15s | %-20s | %-20s\n", "ID", "Type", "Project", "Raw (Snippet)", "Compressed")
	fmt.Println("----------------------------------------------------------------------------------------------------")
	for rows.Next() {
		var id int
		var oType, proj, raw, comp string
		rows.Scan(&id, &oType, &proj, &raw, &comp)

		// Snippet the raw text so it doesn't break the terminal layout
		rawSnippet := raw
		if len(raw) > 20 {
			rawSnippet = raw[:17] + "..."
		}
		compSnippet := comp
		if len(comp) > 20 {
			compSnippet = comp[:17] + "..."
		}

		fmt.Printf("%-4d | %-12s | %-15s | %-20s | %-20s\n", id, oType, proj, rawSnippet, compSnippet)
	}
}

func printSummaries(db *sql.DB) {
	fmt.Println("\n📝 [SUMMARIES]")
	rows, err := db.Query("SELECT id, session_id, project, summary_text FROM summaries")
	if err != nil {
		fmt.Printf("Error fetching summaries: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Printf("%-4s | %-20s | %-20s | %-20s\n", "ID", "SessionID", "Project", "Summary")
	fmt.Println("----------------------------------------------------------------")
	for rows.Next() {
		var id int
		var sid, proj, summ string
		rows.Scan(&id, &sid, &proj, &summ)
		fmt.Printf("%-4d | %-20s | %-20s | %-20s\n", id, sid, proj, summ)
	}
}
