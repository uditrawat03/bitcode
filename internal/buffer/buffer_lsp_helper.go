package buffer

import (
	"path/filepath"
	"strings"

	lsp "github.com/uditrawat03/bitcode/internal/lsp_client"
)

// URI returns the LSP file URI for this buffer
func (b *Buffer) URI() string {
	if b.File == "" {
		return ""
	}
	abs, _ := filepath.Abs(b.File)
	return "file://" + abs
}

func TrimUri(uri string) string {
	return strings.TrimPrefix(uri, "file://")
}

// Position returns the current cursor position as an LSP position
func (b *Buffer) Position() lsp.Position {
	return lsp.Position{
		Line:      b.CursorY,
		Character: b.CursorX,
	}
}

// SelectionRange returns the current selection as an LSP range
// (for CodeAction requests). If no selection, it falls back to cursor.
func (b *Buffer) SelectionRange() lsp.Range {
	start := lsp.Position{Line: b.CursorY, Character: b.CursorX}
	end := start
	// TODO: if your Editor tracks selection (selStartX/selStartY etc.),
	// wire that here instead of always returning just the cursor.
	return lsp.Range{Start: start, End: end}
}
