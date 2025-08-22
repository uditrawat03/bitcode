package lsp

import (
	"context"
	"encoding/json"
	"fmt"
)

// ---- Request Params ----
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type CodeActionContext struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Context      CodeActionContext      `json:"context"`
}

func (s *Client) CodeAction(ctx context.Context, uri string, rng Range) ([]CodeAction, error) {
	params := CodeActionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Range:        rng,
		Context:      CodeActionContext{Diagnostics: []Diagnostic{}},
	}

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
