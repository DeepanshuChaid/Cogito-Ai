package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/injector"
)

type CodexHookInput struct {
	SessionID string `json:"session_id"`
	CWD       string `json:"cwd"`
	Prompt    string `json:"prompt,omitempty"`
}

type CodexHookOutput struct {
	Continue           bool                     `json:"continue"`
	SuppressOutput     bool                     `json:"suppressOutput,omitempty"`
	HookSpecificOutput *CodexHookSpecificOutput `json:"hookSpecificOutput,omitempty"`
	SystemMessage      string                   `json:"systemMessage,omitempty"`
}

type CodexHookSpecificOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "install" {
		runInstall()
		return
	}

	input := CodexHookInput{}
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		input.CWD, _ = os.Getwd()
		input.SessionID = "unknown"
	}
	if strings.TrimSpace(input.CWD) == "" {
		input.CWD, _ = os.Getwd()
	}
	if strings.TrimSpace(input.SessionID) == "" {
		input.SessionID = "unknown"
	}

	context := injector.GenerateContext(input.CWD, input.SessionID)

	output := CodexHookOutput{
		Continue:       true,
		SuppressOutput: true,
		HookSpecificOutput: &CodexHookSpecificOutput{
			HookEventName:     "SessionStart",
			AdditionalContext: context,
		},
		SystemMessage: context,
	}

	jsonOut, err := json.Marshal(output)
	if err != nil {
		fmt.Println(`{"continue":true,"suppressOutput":true}`)
		return
	}
	fmt.Println(string(jsonOut))
}

func runInstall() {
	fmt.Println("Installing Cogito...")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Failed to resolve home directory: %v\n", err)
		os.Exit(1)
	}
	configDir := filepath.Join(homeDir, ".cogito")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("Failed to create config dir: %v\n", err)
		os.Exit(1)
	}

	contextFile := filepath.Join(configDir, "context.md")
	if _, err := os.Stat(contextFile); os.IsNotExist(err) {
		defaultContext := `# Cogito Context

## Instructions
- Be terse and direct
- Drop filler words (just, really, basically, sure, happy to help)
- Drop articles (a, an, the) when possible
- Fragments OK
- Technical accuracy > politeness

## Pattern
[thing] [action] [reason]. [next step].

## Examples
Bad: "Sure! I'd be happy to help you with that. The issue is..."
Good: "Bug in auth middleware. Token expiry check use < not <=. Fix:"

## Exceptions
- Code blocks: unchanged
- Security explanations: normal English
- User says "normal mode": deactivate
`
		if err := os.WriteFile(contextFile, []byte(defaultContext), 0644); err != nil {
			fmt.Printf("Failed to create context file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Created config: ~/.cogito/context.md")
	} else {
		fmt.Println("Config already exists: ~/.cogito/context.md")
	}

	cwd, _ := os.Getwd()
	hooksDir := filepath.Join(cwd, ".codex")

	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		fmt.Printf("Failed to create .codex dir: %v\n", err)
		os.Exit(1)
	}

	hooksFile := filepath.Join(hooksDir, "hooks.json")
	hooksConfig := map[string]interface{}{
		"hooks": map[string]interface{}{
			"SessionStart": []map[string]string{
				{
					"type":    "command",
					"command": resolveHookCommand(),
				},
			},
		},
	}
	hooksContent, err := json.MarshalIndent(hooksConfig, "", "  ")
	if err != nil {
		fmt.Printf("Failed to build hooks.json: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(hooksFile, hooksContent, 0644); err != nil {
		fmt.Printf("Failed to create hooks.json: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created hook: %s/.codex/hooks.json\n", cwd)

	fmt.Println("\nCogito installed successfully.")
	fmt.Println("\nNext steps:")
	fmt.Println("   1. Run 'codex' in this folder")
	fmt.Println("   2. Ask a question to see caveman-style responses")
	fmt.Println("   3. Edit ~/.cogito/context.md to customize instructions")
}

func resolveHookCommand() string {
	if _, err := exec.LookPath("cogito"); err == nil {
		return "cogito"
	}

	execPath, err := os.Executable()
	if err != nil {
		return "cogito"
	}

	execPath = filepath.Clean(execPath)
	tempDir := strings.ToLower(filepath.Clean(os.TempDir()))
	if strings.HasPrefix(strings.ToLower(execPath), tempDir) {
		return "cogito"
	}
	if strings.Contains(execPath, " ") {
		return fmt.Sprintf("\"%s\"", execPath)
	}
	return execPath
}
