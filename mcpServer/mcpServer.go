package mcpServer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/db"
	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/models/schemaModels"
)

type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      *json.RawMessage       `json:"id,omitempty"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

type JSONRPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id"`
	Result  interface{}      `json:"result,omitempty"`
	Error   interface{}      `json:"error,omitempty"`
}

// ServerState keeps track of the current session during the lifecycle
var currentSession *schemaModels.Session

func handleRequest(req JSONRPCRequest) interface{} {
    switch req.Method {
    case "initialize":
        // 1. Get workspace path from params (standard MCP behavior)
        var cwd string
        if _, ok := req.Params["capabilities"].(map[string]interface{}); ok {
            // Note: In a real MCP client, check rootUri or workspaceFolders
            // For now, we fallback to Getwd if the client is simple
            cwd, _ = os.Getwd()
        } else {
            cwd, _ = os.Getwd()
        }

        // 2. Create a unique ID based on Parent Process ID (Codex/Claude process)
        uniqueID := fmt.Sprintf("codex-%d", os.Getppid())

        // 3. Initialize/Recover Session
        session, err := db.InitializeProjectSession(uniqueID, cwd)
        if err == nil {
            currentSession = session
        }

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

	case "initialized":
		// This is a notification in many clients, but some send it with an ID
		return map[string]interface{}{}

	case "tools/list":
		return map[string]interface{}{
			"tools": []map[string]interface{}{
				{
					"name": "get_codebase_map",
					"description": "Lists all files in the project",
					"inputSchema": map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
					},
				},
			},
		}

	case "prompts/list":
		return map[string]interface{}{
			"prompts": []map[string]interface{}{
				{
					"name":        "caveman-review",
					"description": "Ultra-compressed code review",
				},
			},
		}

    case "prompts/get":
        name, _ := req.Params["name"].(string)
        if name == "caveman-review" {
            // --- THE INJECTION ---
            // Fetch observations for this project to give the AI "Lore"
            lore := ""
            if currentSession != nil {
                // You'll need to write this DB function next!
                // observations := db.GetProjectLore(currentSession.ProjectID)
                // lore = formatObservations(observations)
            }

            return map[string]interface{}{
                "messages": []map[string]interface{}{
                    {
                        "role": "system",
                        "content": map[string]interface{}{
                            "type": "text",
                            "text": PROMPT + "\n\n## Project Lore\n" + lore,
                        },
                    },
                },
            }
        }
        return errorResponse(-32601, "prompt not found")

    case "tools/call":
        name, _ := req.Params["name"].(string)
        if name == "get_codebase_map" {
            // Use currentSession.Project if available to ensure we stay in bounds
            root := "."
            if currentSession != nil {
                root = currentSession.Project
            }

            output := ""
            filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
                if err == nil && !info.IsDir() && !strings.HasPrefix(filepath.Base(path), ".") {
                    output += path + "\n"
                }
                return nil
            })

            return map[string]interface{}{
                "content": []map[string]interface{}{
                    {"type": "text", "text": output},
                },
            }
        }
    }
    return errorResponse(-32601, "method not found")
}

func ServeMcp() {
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		var req JSONRPCRequest
		err := json.Unmarshal(scanner.Bytes(), &req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "JSON decode error: %v\n", err)
			continue
		}

		// Notifications (no ID) are handled silently
		if req.ID == nil {
			fmt.Fprintf(os.Stderr, "NOTIFICATION: %s\n", req.Method)
			continue
		}

		result := handleRequest(req)
		resp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
		}

		if m, ok := result.(map[string]interface{}); ok {
			if errVal, exists := m["error"]; exists {
				resp.Error = errVal
			} else {
				resp.Result = m
			}
		} else {
			resp.Result = result
		}

		encoder.Encode(resp)
	}
}

const PROMPT = `# Cogito Review Mode

Ultra-compressed code review comments. Cuts noise from PR feedback while preserving
the actionable signal. Each comment is one line: location, problem, fix. Use when user
says "review this PR", "code review", "review the diff", "/review", or invokes
/caveman-review. Auto-triggers when reviewing pull requests.

---

Write code review comments terse and actionable. One line per finding. Location, problem, fix. No throat-clearing.

## Rules

**Format:** L<line>: <problem>. <fix>. — or <file>:L<line>: ... when reviewing multi-file diffs.

**Severity prefix (optional, when mixed):**
- 🔴 bug: — broken behavior, will cause incident
- 🟡 risk: — works but fragile (race, missing null check, swallowed error)
- 🔵 nit: — style, naming, micro-optim. Author can ignore
- ❓ q: — genuine question, not a suggestion

**Drop:**
- "I noticed that...", "It seems like...", "You might want to consider..."
- "This is just a suggestion but..." — use nit: instead
- "Great work!", "Looks good overall but..." — say it once at the top, not per comment
- Restating what the line does — the reviewer can read the diff
- Hedging ("perhaps", "maybe", "I think") — if unsure use q:

**Keep:**
- Exact line numbers
- Exact symbol/function/variable names in backticks
- Concrete fix, not "consider refactoring this"
- The *why* if the fix isn't obvious from the problem statement

## Examples

❌ "I noticed that on line 42 you're not checking if the user object is null before accessing the email property. This could potentially cause a crash if the user is not found in the database. You might want to add a null check here."

✅ L42: 🔴 bug: user can be null after .find(). Add guard before .email.

❌ "It looks like this function is doing a lot of things and might benefit from being broken up into smaller functions for readability."

✅ L88-140: 🔵 nit: 50-line fn does 4 things. Extract validate/normalize/persist.

❌ "Have you considered what happens if the API returns a 429? I think we should probably handle that case?"

✅ L23: 🟡 risk: no retry on 429. Wrap in withBackoff(3).

## Auto-Clarity

Drop terse mode for: security findings (CVE-class bugs need full explanation + reference), architectural disagreements (need rationale, not just a one-liner), and onboarding contexts where the author is new and needs the "why". In those cases write a normal paragraph, then resume terse for the rest.

## Boundaries

Reviews only — does not write the code fix, does not approve/request-changes, does not run linters. Output the comment(s) ready to paste into the PR. "stop caveman-review" or "normal mode": revert to verbose review style.
`

func errorResponse(code int, msg string) map[string]interface{} {
    return map[string]interface{}{
        "error": map[string]interface{}{
            "code":    code,
            "message": msg,
        },
    }
}
