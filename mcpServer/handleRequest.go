package mcpServer

import (
	"os"
	// "path/filepath"
	// "strings"

	// "github.com/DeepanshuChaid/Cogito-Ai.git/internals/commands"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/commands"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/db"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/models/schemaModels"
)


var currentSession *schemaModels.Session


func handleRequest(req JSONRPCRequest) interface{} {

	switch req.Method {

	//==============================================
	case "initialize":
		cwd, _ := os.Getwd()

		uniqueID := newSessionID()

		session, err := db.InitializeProjectSession(uniqueID, cwd)
		if err == nil {
			currentSession = session
		}

		// go func() {
		// 	commands.BuildMap()
		// }()

		// go commands.BuildMap()


		return map[string]interface{}{
			"protocolVersion": "2025-06-18",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": false,
				},
				"prompts": map[string]interface{}{
					"listChanged": false,
				},
			},
			"serverInfo": map[string]interface{}{
				"name":    "cogito",
				"version": "0.1.0",
			},
		}

	//==============================================
	case "initialized":
		return map[string]interface{}{}

	//==============================================
	case "tools/list":
		return map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name":        "caveman_review",
					"description": "Ultra strict compressed code review",
					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"code": map[string]interface{}{
								"type": "string",
							},
						},
						"required": []string{"code"},
					},
				},
				{
					"name":        "get_codebase_map",
					"description": "Get A full Map of the Codebase with Details like importance and functions flow.",
					"inputSchema": map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
			},
		}

	//==============================================
	case "prompts/list":
		return map[string]interface{}{
			"prompts": []map[string]interface{}{
				{
					"name":        "caveman-review",
					"description": "Ultra-compressed code review",
				},
			},
		}

	//==============================================
	case "prompts/get":
		name, _ := req.Params["name"].(string)

		if name == "caveman-review" {

			lore := ""
			if currentSession != nil {
				// future: fetch observations
			}

			return map[string]interface{}{
				"messages": []map[string]interface{}{
					{
						"role": "system",
						"content": map[string]interface{}{
							"type": "text",
							"text": PROMPT + "\n\n" + lore,
						},
					},
				},
			}
		}

		return errorResponse(-32601, "prompt not found")

	//==============================================
	case "tools/call":
		name, _ := req.Params["name"].(string)

		if name == "caveman_review" {
			arg, ok := req.Params["arguments"].(map[string]interface{})
			if !ok {
				return errorResponse(-32602, "arguments missing")
			}

			code, ok := arg["code"].(string)
			if !ok {
				return errorResponse(-32602, "code missing")
			}

			code = trimInput(code)

			prompt := PROMPT + "\n\nCODE:\n" + code

			result, err := runCaveman(prompt)
			if err != nil {
				return errorResponse(-32603, err.Error())
			}

			return map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": cleanOutput(result)},
				},
			}
		}

		if name == "get_codebase_map" {
			commands.BuildMap(false)

			data, _ := os.ReadFile(".cogito/substrate.txt")
			output := string(data)

			return map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": output},
				},
			}
		}

	}

	return errorResponse(-32601, "method not found")
}
