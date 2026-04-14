package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/injector"
	"github.com/creasty/defaults"
)

// Codex hook input structure
type CodexHookInput struct {
	SessionID string `json:"session_id"`
	CWD       string `json:"cwd"`
	Prompt    string `json:"prompt,omitempty"`
}

// Codex hook output structure
type CodexHookOutput struct {
	Continue       bool   `json:"continue"`
	SystemMessage  string `json:"system_message,omitempty"`
}

func main() {
	// IDK KNOW HOW THIS WORKS BUT ..
	if len(os.Args) > 1 && os.Args[1] == "install" {
		runInstall()
	}


	// Read stdin from Codex hook
	input := CodexHookInput{}
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		// If stdin is empty, use defaults
		input.CWD, _ = os.Getwd()
		input.SessionID = "unknown"
	}

	// Generate context
	context := injector.GenerateContext(input.CWD, input.SessionID)

	// Build output
	output := CodexHookOutput{
		Continue:      true,
		SystemMessage: context,
	}

	// Write to stdout (Codex reads this)
	jsonOut, _ := json.Marshal(output)
	fmt.Println(string(jsonOut))
}

func runInstall() {
	fmt.Println("Installing Cogito...")

	homeDir, _ := os.UserHomeDir()

	//	CREATE A CONFIG DIRECTORY
	configDIr := filepath.Join(homeDir, ".cogito")

	if err := os.Mkdir(configDIr, 0755); err != nil {
		fmt.Printf("❌ Failed to create config dir: %v", err)
		os.Exit(1)
	}

	// CREATE A DEFAULT CONTEXT FILE
	contextFile := filepath.Join(configDIr, "context.md")
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
		if err :=  os.WriteFile(contextFile, []byte(defaultContext), 0644); err != nil {
			fmt.Printf("❌ Failed to Create context file: ", err)
			os.Exit(1)
		}

		fmt.Println("✅ Created Config file")
	} else {
		fmt.Println("Cogito config File already exists")
	}

	cwd, _ := os.Getwd()
	hooksDir := filepath.Join(cwd, ".codex")

	if err := os.Mkdir(hooksDir, 0755); err != nil {
		fmt.Printf("Failed to create.codex dir: ", err)
		os.Exit(1)
	}

	hooksFile := filepath.Join(hooksDir, "hooks.json")
	hooksContent := `{
	"hooks": {
		"SessionStart": [
		{
			"type": "command",
			"command": "cogito"
		}
		]
	}
	}
	`

	if err := os.WriteFile(hooksFile, []byte(hooksContent), 0644); err != nil {
		fmt.Printf("Failed to create hooks.json: %v", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Created hook: %s/.codex/hooks.json\n", cwd)

	// 4. Instructions
	fmt.Println("\n✅ Cogito installed successfully!")
	fmt.Println("\n📍 Next steps:")
	fmt.Println("   1. Run 'codex' in this folder")
	fmt.Println("   2. Ask a question to see caveman-style responses")
	fmt.Println("   3. Edit ~/.cogito/context.md to customize instructions")
	fmt.Println("\n🔧 To install in another project:")
	fmt.Println("   cd <project-folder> && cogito install")
}
