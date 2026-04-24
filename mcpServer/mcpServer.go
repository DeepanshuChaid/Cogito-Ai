package mcpServer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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

var currentSession *schemaModels.Session

// 🔥 SHORT + AGGRESSIVE = WORKS
const CAVEMAN_CORE = `
Terse like caveman. Technical substance exact.
No fluff. No filler. No pleasantries.
Fragments OK. Short sentences.
ALWAYS ACTIVE.
`

func handleRequest(req JSONRPCRequest) interface{} {

	switch req.Method {

	//==============================================
	//==============================================
	//==============================================
	case "initialize":
		cwd, _ := os.Getwd()

		uniqueID := fmt.Sprintf("session-%d", os.Getppid())

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

			// 🔥 SYSTEM-LEVEL INJECTION
			"instructions": CAVEMAN_CORE,
		}

		//==============================================
	//==============================================
	//==============================================
case "initialized":
	return map[string]interface{}{}


	//==============================================
	//==============================================
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
		},
	}

		//==============================================
		//==============================================
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
		//==============================================
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
							"text": CAVEMAN_CORE + "\n\n" + PROMPT + "\n\n" + lore,
						},
					},
				},
			}
		}

		return errorResponse(-32601, "prompt not found")

		//==============================================
		//==============================================
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

			prompt := CAVEMAN_CORE + "\n\n" + PROMPT + "\n\nCODE:\n" + code

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

		// 🔥 FALLBACK: FORCE ALL TEXT THROUGH CAVEMAN
		if name == "" {
			input, _ := req.Params["input"].(string)

			prompt := CAVEMAN_CORE + "\n\n" + PROMPT + "\n\n" + input

			out, err := runCaveman(prompt)
			if err != nil {
				return errorResponse(-32603, err.Error())
			}

			return map[string]interface{}{
				"content": []map[string]interface{}{
					{"type": "text", "text": out},
				},
			}
		}

		if name == "get_codebase_map" {
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
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	encoder := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		var req JSONRPCRequest

		err := json.Unmarshal(scanner.Bytes(), &req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "JSON decode error: %v\n", err)
			continue
		}

		if req.ID == nil {
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

// 🔥 KEEP IT SHORT (important)
const PROMPT = `
Code review mode.

STRICT RULES:
- Max 12 words per line
- Max 5 lines total
- If exceeded → output INVALID

Format: L<line>: <problem>. <fix>.
No fluff.
`

func errorResponse(code int, msg string) map[string]interface{} {
	return map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": msg,
		},
	}
}


func validateOutput(text string) error {
	lines := strings.Split(strings.TrimSpace(text), "\n")

	if len(lines) > 5 {
		return fmt.Errorf("too many lines")
	}

	for _, line := range lines {
		if len(strings.Fields(line)) > 12 {
			return fmt.Errorf("line too long")
		}
	}

	return nil
}

func trimInput(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) > 200 {
		lines = lines[:200]
	}
	return strings.Join(lines, "\n")
}


func runCaveman(prompt string) (string, error) {
	var output string

	for i := 0; i < 2; i++ {
		out, err := callModel(prompt) // <-- YOU implement this
		if err != nil {
			return "", err
		}

		if err := validateOutput(out); err == nil {
			return out, nil
		} else {
			// tighten prompt
			prompt = fmt.Sprintf(`
Fix output. Too verbose.

ERROR: %v

OUTPUT:
%s

Return shorter version only. Follow rules strictly.
`, err, out)

			output = out
		}
	}

	return output, fmt.Errorf("failed to enforce caveman constraints")
}


func callModel(prompt string) (string, error) {
	cmd := exec.Command("codex", "run") // <-- change to your actual CLI

	cmd.Stdin = strings.NewReader(prompt)

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("model error: %v | %s", err, stderr.String())
	}

	return out.String(), nil
}

func cleanOutput(s string) string {
	return strings.TrimSpace(s)
}
