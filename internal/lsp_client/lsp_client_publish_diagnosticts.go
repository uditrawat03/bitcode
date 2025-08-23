package lsp

import (
	"context"
	"encoding/json"
)

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity,omitempty"`
	Code     string `json:"code,omitempty"`
	Source   string `json:"source,omitempty"`
	Message  string `json:"message"`
}

func (s *Client) OnDiagnostics(handler func(uri string, diags []Diagnostic)) {
	s.logger.Printf("[Diagnostics]")

	s.RegisterHandler("textDocument/publishDiagnostics", func(_ context.Context, raw json.RawMessage) {
		var params PublishDiagnosticsParams
		if err := json.Unmarshal(raw, &params); err != nil {
			s.logger.Printf("failed to parse diagnostics: %v", err)
			return
		}
		s.logger.Printf("[Diagnostics] %s -> %+v", params.URI, params.Diagnostics)

		// call user-supplied handler
		if handler != nil {
			handler(params.URI, params.Diagnostics)
		}
	})
}
