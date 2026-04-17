package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/utils/configTui"
)

func Install() {
	configTui.RunConfigTUI()
	fmt.Println("Installing Cogito...")

	execPath, _ := os.Executable()
	execPath, _ = filepath.EvalSymlinks(execPath) // 🔥 important

	cwd, _ := os.Getwd()

	// Create .cogito
	cogitoDir := filepath.Join(cwd, ".cogito")
	os.MkdirAll(cogitoDir, 0755)

	// Create .codex
	hooksDir := filepath.Join(cwd, ".codex")
	os.MkdirAll(hooksDir, 0755)

	// ✅ Quote path (Windows-safe)
	// commandPath := fmt.Sprintf("\"%s\"", execPath)
	commandPath := execPath // ✅ correct

	hooksConfig := map[string]interface{}{
		"hooks": map[string]interface{}{
			"SessionStart": []map[string]string{
				{
					"type":    "command",
					"command": commandPath,
				},
			},
		},
	}

	content, _ := json.MarshalIndent(hooksConfig, "", "  ")
	os.WriteFile(filepath.Join(hooksDir, "hooks.json"), content, 0644)

	fmt.Println("\n✅ Hook installed successfully!")
}
