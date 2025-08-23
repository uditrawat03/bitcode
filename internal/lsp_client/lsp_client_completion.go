package lsp

import (
	"context"
	"encoding/json"
	"fmt"
)

// --- LSP structs ---

type CompletionContext struct {
	TriggerKind      int    `json:"triggerKind"`                // 1 = Invoked, 2 = TriggerCharacter, 3 = TriggerForIncompleteCompletions
	TriggerCharacter string `json:"triggerCharacter,omitempty"` // Character triggering completion
}

type CompletionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Context      *CompletionContext     `json:"context,omitempty"`
}

type CompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind,omitempty"`
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	InsertText    string `json:"insertText,omitempty"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// --- Completion client methods ---

func (c *Client) OnCompletion(handler func(uri string, items []CompletionItem)) {
	c.RegisterHandler("textDocument/completion", func(ctx context.Context, raw json.RawMessage) {
		type Response struct {
			ID     interface{}      `json:"id"`
			Result json.RawMessage  `json:"result"`
			Error  *json.RawMessage `json:"error,omitempty"`
		}
		var resp Response
		if err := json.Unmarshal(raw, &resp); err != nil {
			return
		}
		if resp.Result == nil {
			return
		}

		// Try CompletionList first
		var list CompletionList
		if err := json.Unmarshal(resp.Result, &list); err == nil && len(list.Items) > 0 {
			handler("", list.Items)
			return
		}

		// Try raw []CompletionItem
		var items []CompletionItem
		if err := json.Unmarshal(resp.Result, &items); err == nil && len(items) > 0 {
			handler("", items)
		}
	})
}

func (s *Client) Completion(ctx context.Context, uri string, pos Position, ctxData *CompletionContext) ([]CompletionItem, error) {
	params := CompletionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     pos,
		Context:      ctxData,
	}

	ch, err := s.SendRequest(ctx, "textDocument/completion", params)
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("completion error: %s", resp.Error.Message)
		}

		if resp.Result == nil {
			return nil, nil
		}

		// CompletionList
		var list CompletionList
		if err := json.Unmarshal(*resp.Result, &list); err == nil && len(list.Items) > 0 {
			return list.Items, nil
		}

		// raw []CompletionItem
		var items []CompletionItem
		if err := json.Unmarshal(*resp.Result, &items); err == nil {
			return items, nil
		}

		return nil, fmt.Errorf("invalid completion response format")
	}
}
