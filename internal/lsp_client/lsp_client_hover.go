package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// MarkedString or MarkupContent
type MarkupContent struct {
	Kind  string `json:"kind,omitempty"`
	Value string `json:"value,omitempty"`
}

type HoverResult struct {
	Contents json.RawMessage `json:"contents"`
	Range    *Range          `json:"range,omitempty"`
}

func (s *Client) Hover(ctx context.Context, uri string, pos Position) (string, error) {
	params := map[string]interface{}{
		"textDocument": map[string]string{"uri": uri},
		"position":     pos,
	}

	ch, err := s.SendRequest(ctx, "textDocument/hover", params)
	if err != nil {
		return "", err
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()

	case resp := <-ch:
		if resp.Error != nil {
			return "", fmt.Errorf("hover error: %s", resp.Error.Message)
		}
		if resp.Result == nil {
			return "", nil
		}

		var hover HoverResult
		if err := json.Unmarshal(*resp.Result, &hover); err != nil {
			return "", fmt.Errorf("invalid hover result: %w", err)
		}

		// contents can be string | MarkedString | []MarkedString | MarkupContent
		var (
			contentStr string
			tryErr     error
		)

		// Try string
		var sVal string
		if err := json.Unmarshal(hover.Contents, &sVal); err == nil {
			return sVal, nil
		}

		// Try MarkupContent
		var markup MarkupContent
		if err := json.Unmarshal(hover.Contents, &markup); err == nil && markup.Value != "" {
			return markup.Value, nil
		}

		// Try []MarkedString or []string
		var arr []json.RawMessage
		if err := json.Unmarshal(hover.Contents, &arr); err == nil {
			var parts []string
			for _, elem := range arr {
				var str string
				if err := json.Unmarshal(elem, &str); err == nil {
					parts = append(parts, str)
					continue
				}
				var ms MarkupContent
				if err := json.Unmarshal(elem, &ms); err == nil && ms.Value != "" {
					parts = append(parts, ms.Value)
				}
			}
			if len(parts) > 0 {
				return strings.Join(parts, "\n"), nil
			}
		}

		return contentStr, tryErr
	}
}
