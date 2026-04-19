package main

import (
    "encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// SaveObservation logs tool use to file
func SaveObservation(tool string, input, output string) {
    data, _ := json.MarshalIndent(map[string]interface{}{
        "time":   time.Now().Unix(),
        "tool":   tool,
        "input":  input,
        "output": output,
    }, "", "  ")
    os.WriteFile(".cogito/last_action.json", data, 0644)
}

// GET /mcp/tools — required by MCP
func toolsHandler(w http.ResponseWriter, r *http.Request) {
    tools := []map[string]interface{}{
        {
            "name":        "on_tool_use",
            "description": "Internal: receive tool usage from Codex",
            "input_schema": map[string]interface{}{
                "type":       "object",
                "properties": map[string]map[string]string{},
            },
        },
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "tools": tools,
    })
}

// POST /mcp/invoke — Codex sends events here
func invokeHandler(w http.ResponseWriter, r *http.Request) {
    var body map[string]interface{}
    json.NewDecoder(r.Body).Decode(&body)

    toolName := body["name"].(string)
    args := body["arguments"].(map[string]interface{})

    input, _ := json.Marshal(args["input"])
    output, _ := json.Marshal(args["output"])

    // Log to file
	SaveObservation(toolName, string(input), string(output))

    // Respond
	w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"content": "Logged",
    })
}

func main() {
    os.MkdirAll(".cogito", 0755)

    // Log every request
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        log.Printf("🚨 %s %s", r.Method, r.URL.Path)

        if r.URL.Path == "/mcp/tools" {
            toolsHandler(w, r)
            return
}
        if r.URL.Path == "/mcp/invoke" {
            invokeHandler(w, r)
            return
}

        log.Printf("❌ 404: %s", r.URL.Path)
        http.Error(w, "not found", http.StatusNotFound)
    })

    log.Println("✅ MCP Server: Listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
