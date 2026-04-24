package mcpServer

import "encoding/json"

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
