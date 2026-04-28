package mcpServer

import (
	"fmt"
	"hash/fnv"
	"os"
	"strings"
	// "path/filepath"
	// "strings"

	// "github.com/DeepanshuChaid/Cogito-Ai.git/internals/commands"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/commands"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/db"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/models/schemaModels"
)

var currentSession *schemaModels.Session

const maxObservationMemoryWords = 25
const maxObservationFactsWords = 10

func shortHash(s string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum64())
}

func handleRequest(req JSONRPCRequest) interface{} {
	// #region agent log
	writeDebugLog(
		"run1",
		"H7",
		"mcpServer/handleRequest.go:handleRequest",
		"entered handleRequest",
		map[string]interface{}{
			"method": req.Method,
		},
	)
	// #endregion

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
					"name":        "create_observation",
					"description": "Store one short durable engineering memory. Use direct caveman style.",

					"inputSchema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"memory": map[string]interface{}{
								"type":        "string",
								"description": "Required. Direct text, <=25 words. Format: change + impact.",
							},
							"facts": map[string]interface{}{
								"type":        "string",
								"description": "Optional. Direct text, <=10 words. Stable facts only.",
							},
						},
						"required": []string{
							"memory",
						},
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

		if name == "create_observation" {
			arg, ok := req.Params["arguments"].(map[string]interface{})
			if !ok {
				return errorResponse(-32602, "arguments missing")
			}

			memoryText, ok := arg["memory"].(string)
			if !ok || memoryText == "" {
				return errorResponse(-32602, "memoryText missing")
			}

			var fact string
			fact, _ = arg["facts"].(string)
			memoryText = strings.TrimSpace(memoryText)
			fact = strings.TrimSpace(fact)

			memoryWords := len(strings.Fields(memoryText))
			factsWords := len(strings.Fields(fact))
			if memoryWords > maxObservationMemoryWords {
				return errorResponse(-32602, fmt.Sprintf("memory too long: max %d words", maxObservationMemoryWords))
			}
			if factsWords > maxObservationFactsWords {
				return errorResponse(-32602, fmt.Sprintf("facts too long: max %d words", maxObservationFactsWords))
			}

			cwd, _ := os.Getwd()

			// #region agent log
			writeDebugLog(
				"run1",
				"H1",
				"mcpServer/handleRequest.go:create_observation",
				"create_observation received",
				map[string]interface{}{
					"memoryHash":  shortHash(memoryText),
					"factsHash":   shortHash(fact),
					"memoryWords": memoryWords,
					"factsWords":  factsWords,
					"project":     cwd,
				},
			)
			// #endregion

			if currentSession == nil {
				return errorResponse(-32602, "no active session")
			}

			isDuplicate, matchedMemory, score, dupErr := db.IsDuplicateObservation(cwd, memoryText, fact, 30)
			if dupErr != nil {
				return errorResponse(-32603, dupErr.Error())
			}

			// #region agent log
			writeDebugLog(
				"run1",
				"H2",
				"mcpServer/handleRequest.go:create_observation",
				"duplicate-check result",
				map[string]interface{}{
					"isDuplicate":       isDuplicate,
					"similarityScore":   score,
					"matchedMemoryHash": shortHash(matchedMemory),
				},
			)
			// #endregion

			if isDuplicate {
				return map[string]interface{}{
					"content": []map[string]interface{}{
						{"type": "text", "text": "Skipped duplicate observation"},
					},
				}
			}

			err := db.CreateObservation(currentSession.SessionID, cwd, memoryText, fact)
			if err != nil {
				return errorResponse(-32603, err.Error())
			}

			// #region agent log
			writeDebugLog(
				"run1",
				"H3",
				"mcpServer/handleRequest.go:create_observation",
				"observation inserted",
				map[string]interface{}{
					"sessionID": currentSession.SessionID,
					"project":   cwd,
				},
			)
			// #endregion

			// importance := 5
			// if val, ok := arg["importance"].(float64); ok {
			// 	importance = int(val)
			// }

			return map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": "Observations Saved Successfully!"},
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
