package buffer

import lsp "github.com/uditrawat03/bitcode/internal/lsp_client"

// ---------------- Editing -----------------

func (b *Buffer) InsertRune(ch rune) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.pushUndoLocked()

	if b.CursorY >= len(b.Content) {
		b.Content = append(b.Content, []rune{})
	}
	line := b.Content[b.CursorY]
	if b.CursorX > len(line) {
		b.CursorX = len(line)
	}

	b.Content[b.CursorY] = append(line[:b.CursorX], append([]rune{ch}, line[b.CursorX:]...)...)
	b.CursorX++
	b.scheduleDidChange()

	go b.RequestCompletions(lsp.Position{Line: b.CursorY, Character: b.CursorX})
}

func (b *Buffer) DeleteRune() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.CursorY >= len(b.Content) {
		return
	}

	b.pushUndoLocked()

	line := b.Content[b.CursorY]

	if b.CursorX == 0 {
		if b.CursorY > 0 {
			prev := b.Content[b.CursorY-1]
			b.CursorX = len(prev)
			b.Content[b.CursorY-1] = append(prev, line...)
			b.Content = append(b.Content[:b.CursorY], b.Content[b.CursorY+1:]...)
			b.CursorY--
		}
	} else {
		b.Content[b.CursorY] = append(line[:b.CursorX-1], line[b.CursorX:]...)
		b.CursorX--
	}
	b.scheduleDidChange()
}

func (b *Buffer) InsertLine() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.pushUndoLocked()

	if b.CursorY >= len(b.Content) {
		b.Content = append(b.Content, []rune{})
	} else {
		line := b.Content[b.CursorY]
		left, right := line[:b.CursorX], line[b.CursorX:]
		b.Content[b.CursorY] = left
		b.Content = append(b.Content[:b.CursorY+1], append([][]rune{right}, b.Content[b.CursorY+1:]...)...)
	}
	b.CursorY++
	b.CursorX = 0
	b.scheduleDidChange()
}

func (b *Buffer) SetLine(y int, text string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if y >= 0 && y < len(b.Content) {
		b.pushUndoLocked()
		b.Content[y] = []rune(text)
		b.scheduleDidChange()
	}
}

func (b *Buffer) Line(y int) string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if y < 0 || y >= len(b.Content) {
		return ""
	}
	return string(b.Content[y])
}

func (b *Buffer) Lines() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	lines := make([]string, len(b.Content))
	for i, l := range b.Content {
		lines[i] = string(l)
	}
	return lines
}

func (b *Buffer) Undo() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.undoStack) == 0 {
		return
	}
	// push current to redo
	cur := snapshot{Content: cloneLines(b.Content), CursorX: b.CursorX, CursorY: b.CursorY}
	b.redoStack = append(b.redoStack, cur)

	// pop from undo
	last := b.undoStack[len(b.undoStack)-1]
	b.undoStack = b.undoStack[:len(b.undoStack)-1]

	b.Content = cloneLines(last.Content)
	b.CursorX, b.CursorY = last.CursorX, last.CursorY
	b.scheduleDidChange()
}

func (b *Buffer) Redo() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.redoStack) == 0 {
		return
	}
	// push current to undo
	cur := snapshot{Content: cloneLines(b.Content), CursorX: b.CursorX, CursorY: b.CursorY}
	b.undoStack = append(b.undoStack, cur)

	// pop from redo
	last := b.redoStack[len(b.redoStack)-1]
	b.redoStack = b.redoStack[:len(b.redoStack)-1]

	b.Content = cloneLines(last.Content)
	b.CursorX, b.CursorY = last.CursorX, last.CursorY
	b.scheduleDidChange()
}

// pushUndoLocked records the current state BEFORE a mutation.
// Call with b.mu already locked.
func (b *Buffer) pushUndoLocked() {
	shot := snapshot{
		Content: cloneLines(b.Content),
		CursorX: b.CursorX,
		CursorY: b.CursorY,
	}
	b.undoStack = append(b.undoStack, shot)
	// cap history
	if b.maxHistory > 0 && len(b.undoStack) > b.maxHistory {
		// drop oldest
		copy(b.undoStack[0:], b.undoStack[1:])
		b.undoStack = b.undoStack[:b.maxHistory]
	}
	// invalidate redo chain on new edit
	b.redoStack = b.redoStack[:0]
}

func cloneLines(src [][]rune) [][]rune {
	dst := make([][]rune, len(src))
	for i := range src {
		line := make([]rune, len(src[i]))
		copy(line, src[i])
		dst[i] = line
	}
	return dst
}
