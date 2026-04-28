package mcpServer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/DeepanshuChaid/Cogito-Ai.git/internals/db"
)



func ServeMcp() {
	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	encoder := json.NewEncoder(os.Stdout)

	signalChan := make(chan os.Signal, 1)

	signal.Notify(
		signalChan,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	go func () {
		<- signalChan

		db.CompleteSession(currentSession.SessionID)

		os.Exit(0)
	}()


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

