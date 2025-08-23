package buffer

import (
	"path/filepath"
	"time"

	lsp "github.com/uditrawat03/bitcode/internal/lsp_client"
)

// ---------------- Utilities -----------------

func (b *Buffer) ContentAsString() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]rune, 0, 1024)
	for i, line := range b.Content {
		out = append(out, line...)
		if i < len(b.Content)-1 {
			out = append(out, '\n')
		}
	}
	return string(out)
}

func (b *Buffer) scheduleDidChange() {
	if b.LspClient == nil || b.File == "" {
		return
	}

	b.debounceMu.Lock()
	defer b.debounceMu.Unlock()

	if b.debounce != nil {
		b.debounce.Stop()
	}

	b.debounce = time.AfterFunc(150*time.Millisecond, func() {
		b.mu.RLock()
		text := b.contentAsStringLocked()
		pos := lsp.Position{Line: b.CursorY, Character: b.CursorX}
		b.mu.RUnlock()

		b.LspClient.SendDidChange(b.URI(), text)
		b.RequestCompletions(pos)
	})
}

func (b *Buffer) contentAsStringLocked() string {
	out := make([]rune, 0, 1024)
	for i, line := range b.Content {
		out = append(out, line...)
		if i < len(b.Content)-1 {
			out = append(out, '\n')
		}
	}
	return string(out)
}

func (b *Buffer) ParentFolder() string {
	if b.File == "" {
		return "."
	}
	return filepath.Dir(b.File)
}

func (b *Buffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pushUndoLocked()
	b.Content, b.CursorX, b.CursorY = [][]rune{{}}, 0, 0
	b.scheduleDidChange()
}
