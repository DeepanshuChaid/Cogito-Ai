package mcpServer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

)



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

