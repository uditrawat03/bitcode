package lsp

import "encoding/json"

type Request struct {
	RPC    string          `json:"jsonrpc"`
	ID     json.RawMessage `json:"id,omitempty"`
	Method string          `json:"method"`
}

type Response struct {
	RPC    string          `json:"jsonrpc"`
	ID     json.RawMessage `json:"id,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *ResponseError  `json:"error,omitempty"`
}

type Notification struct {
	RPC    string          `json:"jsonrpc"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

type ResponseError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}
