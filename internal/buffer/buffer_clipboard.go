package buffer

// ---------------- Selection & Clipboard -----------------

func normalizeSelection(selStartX, selStartY, selEndX, selEndY int) (int, int, int, int) {
	if selStartY > selEndY || (selStartY == selEndY && selStartX > selEndX) {
		return selEndX, selEndY, selStartX, selStartY
	}
	return selStartX, selStartY, selEndX, selEndY
}

func (b *Buffer) CopyAllOrSelection(selStartX, selStartY, selEndX, selEndY int, all bool) []rune {
	if all {
		b.mu.RLock()
		defer b.mu.RUnlock()
		var text []rune
		for i, line := range b.Content {
			text = append(text, line...)
			if i < len(b.Content)-1 {
				text = append(text, '\n')
			}
		}
		return text
	}
	return b.CopySelection(selStartX, selStartY, selEndX, selEndY)
}

func (b *Buffer) DeleteSelectionOrAll(selStartX, selStartY, selEndX, selEndY int, all bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if all {
		b.Content, b.CursorX, b.CursorY = [][]rune{{}}, 0, 0
		return
	}
	b.DeleteSelection(selStartX, selStartY, selEndX, selEndY)
}

func (b *Buffer) DeleteAtCursor(cursorX, cursorY, selStartY, selEndY int, selecting bool) {
	if selecting {
		b.DeleteSelection(cursorX, selStartY, cursorY, selEndY)
	} else if cursorY < len(b.Content) {
		line := b.Content[cursorY]
		if cursorX < len(line) {
			b.Content[cursorY] = append(line[:cursorX], line[cursorX+1:]...)
		} else if cursorY < len(b.Content)-1 {
			b.Content[cursorY] = append(b.Content[cursorY], b.Content[cursorY+1]...)
			b.Content = append(b.Content[:cursorY+1], b.Content[cursorY+2:]...)
		}
	}
}

// func (b *Buffer) DeleteSelection(selStartX, selStartY, selEndX, selEndY int) {
// 	b.mu.Lock()
// 	defer b.mu.Unlock()

// 	b.pushUndoLocked()

// 	sx, sy, ex, ey := normalizeSelection(selStartX, selStartY, selEndX, selEndY)
// 	if sy == ey {
// 		line := b.Content[sy]
// 		b.Content[sy] = append(line[:sx], line[ex:]...)
// 	} else {
// 		startLine := b.Content[sy][:sx]
// 		endLine := b.Content[ey][ex:]
// 		b.Content[sy] = append(startLine, endLine...)
// 		b.Content = append(b.Content[:sy+1], b.Content[ey+1:]...)
// 	}
// 	b.CursorY, b.CursorX = sy, sx
// 	b.scheduleDidChange()
// }

func (b *Buffer) DeleteSelection(sx, sy, ex, ey int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.Content) == 0 {
		return
	}

	b.pushUndoLocked()

	// Clamp Y coordinates
	if sy < 0 {
		sy = 0
	} else if sy >= len(b.Content) {
		sy = len(b.Content) - 1
	}

	if ey < 0 {
		ey = 0
	} else if ey >= len(b.Content) {
		ey = len(b.Content) - 1
	}

	// Clamp X coordinates
	startLine := b.Content[sy]
	if sx < 0 {
		sx = 0
	} else if sx > len(startLine) {
		sx = len(startLine)
	}

	endLine := b.Content[ey]
	if ex < 0 {
		ex = 0
	} else if ex > len(endLine) {
		ex = len(endLine)
	}

	// Normalize selection: sy <= ey, sx <= ex for single line
	if sy > ey || (sy == ey && sx > ex) {
		sx, ex = ex, sx
		sy, ey = ey, sy
	}

	if sy == ey {
		// Single-line deletion
		line := b.Content[sy]
		b.Content[sy] = append(line[:sx], line[ex:]...)
	} else {
		// Multi-line deletion
		b.Content[sy] = append(b.Content[sy][:sx], b.Content[ey][ex:]...)
		if ey+1 < len(b.Content) {
			b.Content = append(b.Content[:sy+1], b.Content[ey+1:]...)
		} else {
			b.Content = b.Content[:sy+1]
		}
	}

	// Move cursor to start of deleted selection
	b.CursorX, b.CursorY = sx, sy
	b.scheduleDidChange()
}

func (b *Buffer) CopySelection(selStartX, selStartY, selEndX, selEndY int) []rune {
	b.mu.RLock()
	defer b.mu.RUnlock()

	sx, sy, ex, ey := normalizeSelection(selStartX, selStartY, selEndX, selEndY)
	var copied []rune
	for y := sy; y <= ey; y++ {
		line := b.Content[y]
		switch {
		case y == sy && y == ey:
			copied = append(copied, line[sx:ex]...)
		case y == sy:
			copied = append(copied, line[sx:]...)
			copied = append(copied, '\n')
		case y == ey:
			copied = append(copied, line[:ex]...)
		default:
			copied = append(copied, line...)
			copied = append(copied, '\n')
		}
	}
	return copied
}

func (b *Buffer) CutSelection(selStartX, selStartY, selEndX, selEndY int) []rune {
	text := b.CopySelection(selStartX, selStartY, selEndX, selEndY)
	b.DeleteSelection(selStartX, selStartY, selEndX, selEndY)
	return text
}

func (b *Buffer) PasteClipboard(text []rune) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.pushUndoLocked()

	if b.CursorY >= len(b.Content) {
		b.Content = append(b.Content, []rune{})
	}
	lines := [][]rune{}
	curr := []rune{}
	for _, r := range text {
		if r == '\n' {
			lines = append(lines, curr)
			curr = []rune{}
		} else {
			curr = append(curr, r)
		}
	}
	lines = append(lines, curr)

	line := b.Content[b.CursorY]
	before, after := line[:b.CursorX], line[b.CursorX:]

	b.Content[b.CursorY] = append(before, lines[0]...)
	if len(lines) > 1 {
		b.Content = append(b.Content[:b.CursorY+1], append(lines[1:], b.Content[b.CursorY+1:]...)...)
	}
	lastLine := b.CursorY + len(lines) - 1
	b.Content[lastLine] = append(b.Content[lastLine], after...)
	b.CursorY, b.CursorX = lastLine, len(lines[len(lines)-1])

	b.scheduleDidChange() // 🔑 notify LSP after paste
}
