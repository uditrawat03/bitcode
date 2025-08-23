package lsp

import (
	"context"
	"encoding/json"
	"fmt"
)

type CodeAction struct {
	Title       string         `json:"title"`
	Kind        string         `json:"kind,omitempty"`
	Diagnostics []Diagnostic   `json:"diagnostics,omitempty"`
	Edit        *WorkspaceEdit `json:"edit,omitempty"`
	Command     *Command       `json:"command,omitempty"`
}

type CodeActionContext struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Context      CodeActionContext      `json:"context"`
}

func (s *Client) OnCodeAction(handler func(uri string, actions []CodeAction)) {
	s.RegisterHandler("textDocument/codeAction", func(ctx context.Context, raw json.RawMessage) {
		// LSP responses can be wrapped in { "jsonrpc": "2.0", "id": 1, "result": [...] }
		type Response struct {
			ID     interface{}      `json:"id"`
			Result []CodeAction     `json:"result"`
			Error  *json.RawMessage `json:"error,omitempty"`
		}

		var resp Response
		if err := json.Unmarshal(raw, &resp); err != nil {
			return
		}

		if resp.Result != nil {
			// For now we don’t have uri info in the response; you may store it in a request map
			handler("", resp.Result)
		}
	})
}

func (s *Client) CodeAction(ctx context.Context, uri string, rng Range, diags []Diagnostic) ([]CodeAction, error) {
	params := CodeActionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Range:        rng,
		Context:      CodeActionContext{Diagnostics: diags},
	}

	s.logger.Println("[CodeAction]", params)

	ch, err := s.SendRequest(ctx, "textDocument/codeAction", params)
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("codeAction error: %s", resp.Error.Message)
		}

		if resp.Result == nil {
			return nil, nil
		}

		var actions []CodeAction
		if err := json.Unmarshal(*resp.Result, &actions); err != nil {
			return nil, fmt.Errorf("failed to decode code actions: %w", err)
		}
		return actions, nil
	}
}
