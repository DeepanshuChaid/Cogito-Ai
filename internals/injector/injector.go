package injector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GenerateContext(cwd, sessionID string) string {
	// Try to load custom context file
	contextFile := strings.TrimSpace(os.Getenv("COGITO_CONTEXT_FILE"))
	if contextFile == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			contextFile = filepath.Join(homeDir, ".cogito", "context.md")
		}
	}

	if contextFile != "" {
		if content, err := os.ReadFile(contextFile); err == nil {
			return string(content)
		}
	}

	if strings.TrimSpace(cwd) == "" {
		if wd, err := os.Getwd(); err == nil {
			cwd = wd
		}
	}
	if strings.TrimSpace(sessionID) == "" {
		sessionID = "unknown"
	}

	// Default context
	return fmt.Sprintf(`# Cogito Context

## Session
- ID: %s
- Directory: %s

## Instructions
- Be terse and direct
- Drop filler words
- Technical accuracy > politeness
- Code blocks unchanged

## Memory
Memory features coming soon.

## Project Map
Graphify integration coming soon.
`, sessionID, cwd)
}
