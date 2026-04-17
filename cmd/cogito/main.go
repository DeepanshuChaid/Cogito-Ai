package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	// "path/filepath"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/commands"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/config"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/db"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/injector"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/utils/welcomeUi"
)
// NOTE : USE OS.EXIT(1) ISNTEAD OF RETURN AND TRY TO USE SOFT FAILURE INSTEAD OF HARD FAILURE BECAUSE EXIT SHUTS DOWN THE ENTIRE PROGRAM BUT RETURN ONLY CLOSES THE FUNCTION IT IS IN.
// DEFER DOES NOT WORK WITH OS.EXIT(1) BTW
// EXIT(1) REPRESENST FAILURE AND EXIT(0) SIGNALS SUCCESS

func main() {
	if err := db.InitDB(); err != nil {
        fmt.Fprintf(os.Stderr, "Critical Error: Could not initialize DB: %v\n", err)
        os.Exit(1)
    }

	// This will run when main() returns (after handleHook finishes)
	defer func() {
		if err := db.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing DB: %v\n", err)
		}
	}()

	if len(os.Args) > 1 {
		switch os.Args[1] {

		case "install":
			commands.Install()
			return

		case "config":
			commands.HandleConfig()
			return

		case "--help":
			commands.Help()
			return

		case "uninstall":
			commands.Uninstall()
			return

		case "--version", "-v":
			commands.Version()
			return

		case "compress":
			fmt.Println("Janitor is coming soon... (Step 3)")
			return

		case "test":
            commands.Test()

		default:
			commands.Unkown(os.Args[1])
			os.Exit(1)
		}
	}

	// handleHook()
}

func handleHook() {
    stat, _ := os.Stdin.Stat()
    if (stat.Mode() & os.ModeCharDevice) != 0 {
        welcomeUi.ShowWelcomeUI()
        return
    }

    // 1. Read Stdin ONCE. This drains the stream.
    rawInput, err := io.ReadAll(os.Stdin)
    if err != nil {
        fmt.Fprintf(os.Stderr, "DEBUG: Read Stdin Failed: %v\n", err)
        os.Exit(1)
    }

    // 2. Clean the input (BOM and Newlines)
    cleaned := bytes.TrimPrefix(rawInput, []byte("\xef\xbb\xbf"))
    cleaned = bytes.ReplaceAll(cleaned, []byte("\r"), []byte(""))
    cleaned = bytes.ReplaceAll(cleaned, []byte("\n"), []byte(""))

    var input struct {
        CWD    string `json:"cwd"`
        Prompt string `json:"prompt"`
    }

    // 3. Unmarshal the cleaned bytes
    if err := json.Unmarshal(cleaned, &input); err != nil {
        fmt.Fprintf(os.Stderr, "DEBUG: JSON Decode Failed: %v\nRaw: %s\n", err, string(cleaned))
		os.Exit(1)
    }

    cfg, err := config.Load()
    if err != nil || !cfg.Enabled {
        os.Exit(1) // Silent exit if disabled
    }

    // 4. Fetch memories
    memoriesRaw, err := db.GetRelevantObservations(input.CWD, 10)
    if err != nil {
        fmt.Fprintf(os.Stderr, "DEBUG: DB Error: %v\n", err)
        // Don't exit! Just continue without memories.
    }

    var memTexts []string
    for _, m := range memoriesRaw {
        // Use Compressed if it exists, otherwise fallback to Raw
        content := m.CompressedText
        if content == "" {
            content = m.RawText
        }
        memTexts = append(memTexts, fmt.Sprintf("[%s] %s", m.ObservationType, content))
    }

    // 5. Inject and Output
    context := injector.BuildFinalPrompt(input.Prompt, memTexts, cfg)

    output := map[string]interface{}{
        "continue":       true,
        "supress_output": true, // Matches your original struct tag
        "system_message": context,
    }

    jsonOut, _ := json.Marshal(output)
    fmt.Println(string(jsonOut))
}

