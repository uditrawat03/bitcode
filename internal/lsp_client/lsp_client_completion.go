package lsp

import (
	"context"
	"encoding/json"
	"fmt"
)

// CompletionList follows the LSP spec for completion responses
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete,omitempty"`
	Items        []CompletionItem `json:"items"`
}

func (s *Client) Completion(ctx context.Context, uri string, pos Position) ([]CompletionItem, error) {
	params := struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
		Position Position `json:"position"`
	}{
		TextDocument: struct {
			URI string `json:"uri"`
		}{URI: uri},
		Position: pos,
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

		// Try CompletionList first
		var list CompletionList
		if err := json.Unmarshal(*resp.Result, &list); err == nil && len(list.Items) > 0 {
			return list.Items, nil
		}

		// Otherwise, try a raw []CompletionItem
		var items []CompletionItem
		if err := json.Unmarshal(*resp.Result, &items); err == nil {
			return items, nil
		}

		return nil, fmt.Errorf("invalid completion response format")
	}
}
