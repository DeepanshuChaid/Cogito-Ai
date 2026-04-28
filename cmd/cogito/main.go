package main

import (
	"fmt"
	"os"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/commands"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/db"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/utils/welcomeUi"
	"github.com/DeepanshuChaid/Cogito-Ai.git/mcpServer"
)

func main() {
	if err := db.InitDB(); err != nil {
		fmt.Fprintf(os.Stderr, "Critical Error: (I MAY ACT NON CHALANT BUT I AM KINDA JUST A BITCH) Could not initialize DB: %v\n", err)
		os.Exit(1)
	}

	// DO NOT PRINT ANYTHING TO STDOUT HERE.
	// Codex expects pure JSON output only.

	if len(os.Args) > 1 {
		switch os.Args[1] {

		case "install":
			commands.Install()
			return

		case "build-map":
			commands.BuildMap(true)
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

		case "serve-mcp":
			mcpServer.ServeMcp()
			return


		default:
			commands.Unknown(os.Args[1])
			return
		}
	}

	welcomeUi.ShowWelcomeUI()
}

