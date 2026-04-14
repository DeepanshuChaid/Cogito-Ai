package injector

import (
	"fmt"
	"os"
	"path/filepath"
)

func GenerateContext(cwd, sessionID string) string {
	// Try to load custom context file
	contextFile := os.Getenv("COGITO_CONTEXT_FILE")
	if contextFile == "" {
		contextFile = filepath.Join(os.Getenv("HOME"), ".cogito", "context.md")
	}

	if content, err := os.ReadFile(contextFile); err == nil {
		return string(content)
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
