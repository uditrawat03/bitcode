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

// ---------------- LSP Features -----------------

func (b *Buffer) UpdateDiagnostics(diags []lsp.Diagnostic) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.Diagnostics = diags
}

func (b *Buffer) RequestHover(pos lsp.Position) error {
	if b.LspClient == nil {
		return nil
	}
	hover, err := b.LspClient.Hover(b.ctx, b.URI(), pos)
	if err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.HoverInfo = hover
	return nil
}

func (b *Buffer) RequestCompletions(pos lsp.Position) error {
	if b.LspClient == nil {
		return nil
	}

	ctxData := &lsp.CompletionContext{
		TriggerKind: 1, // manual
	}

	items, err := b.LspClient.Completion(b.ctx, b.URI(), pos, ctxData)
	if err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Completions = items
	return nil
}

func (b *Buffer) RequestCodeActions(rng lsp.Range, diags []lsp.Diagnostic) error {
	if b.LspClient == nil {
		return nil
	}
	actions, err := b.LspClient.CodeAction(b.ctx, b.URI(), rng, diags)
	if err != nil {
		return err
	}

	b.mu.Lock()
	b.CodeActions = actions
	b.mu.Unlock()

	return nil
}

func (b *Buffer) GetCodeActions() []lsp.CodeAction {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return append([]lsp.CodeAction{}, b.CodeActions...) // return copy
}
