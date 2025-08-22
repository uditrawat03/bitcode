package lsp

import (
	"context"
	"encoding/json"
)

// PublishDiagnosticsParams matches LSP spec
type PublishDiagnosticsParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Diagnostics  []Diagnostic           `json:"diagnostics"`
}

func (s *Client) OnDiagnostics(handler func(uri string, diags []Diagnostic)) {
	s.RegisterHandler("textDocument/publishDiagnostics", func(_ context.Context, raw json.RawMessage) {
		var params PublishDiagnosticsParams
		if err := json.Unmarshal(raw, &params); err == nil {
			handler(params.TextDocument.URI, params.Diagnostics)
		}
	})
}
