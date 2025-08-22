package buffer

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type Buffer struct {
	Content [][]rune
	File    string
	CursorX int
	CursorY int
	mu      sync.RWMutex
}

func NewBuffer(path string) *Buffer {
	buf := &Buffer{File: path}
	if path != "" {
		buf.Load()
	} else {
		buf.Content = [][]rune{{}}
	}
	return buf
}

func (b *Buffer) Load() {
	file, err := os.Open(b.File)
	if err != nil {
		log.Printf("Unable to open file: %v", err)
		b.Content = [][]rune{{}}
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	b.Content = [][]rune{}
	for scanner.Scan() {
		b.Content = append(b.Content, []rune(scanner.Text()))
	}
	if len(b.Content) == 0 {
		b.Content = [][]rune{{}}
	}
	b.CursorX, b.CursorY = 0, 0
}

func (b *Buffer) Save() {
	if b.File == "" {
		log.Println("No file path specified.")
		return
	}

	file, err := os.Create(b.File)
	if err != nil {
		log.Printf("Error saving file: %v", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range b.Content {
		_, _ = writer.WriteString(string(line) + "\n")
	}
	writer.Flush()
}

// --- Editing helpers ---

func (b *Buffer) InsertRune(ch rune) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.CursorY >= len(b.Content) {
		b.Content = append(b.Content, []rune{})
	}
	line := b.Content[b.CursorY]

	if b.CursorX > len(line) {
		b.CursorX = len(line)
	}

	// insert at cursor
	newLine := append(line[:b.CursorX], append([]rune{ch}, line[b.CursorX:]...)...)
	b.Content[b.CursorY] = newLine
	b.CursorX++
}

func (b *Buffer) DeleteRune() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.CursorY >= len(b.Content) {
		return
	}
	line := b.Content[b.CursorY]

	if b.CursorX == 0 {
		// join with previous line
		if b.CursorY > 0 {
			prev := b.Content[b.CursorY-1]
			b.Content[b.CursorY-1] = append(prev, line...)
			b.Content = append(b.Content[:b.CursorY], b.Content[b.CursorY+1:]...)
			b.CursorY--
			b.CursorX = len(prev)
		}
		return
	}

	// delete before cursor
	newLine := append(line[:b.CursorX-1], line[b.CursorX:]...)
	b.Content[b.CursorY] = newLine
	b.CursorX--
}

func (b *Buffer) InsertLine() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.CursorY >= len(b.Content) {
		b.Content = append(b.Content, []rune{})
		b.CursorY = len(b.Content) - 1
		b.CursorX = 0
		return
	}

	line := b.Content[b.CursorY]
	left := line[:b.CursorX]
	right := line[b.CursorX:]

	// replace current line with left, insert right below
	b.Content[b.CursorY] = left
	b.Content = append(b.Content[:b.CursorY+1], append([][]rune{right}, b.Content[b.CursorY+1:]...)...)

	b.CursorY++
	b.CursorX = 0
}

func (b *Buffer) SetLine(y int, text string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if y < 0 || y >= len(b.Content) {
		return
	}
	b.Content[y] = []rune(text)
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

// ---------------- BufferManager -----------------

type BufferManager struct {
	buffers map[string]*Buffer
	active  *Buffer
	mu      sync.RWMutex
}

func NewBufferManager() *BufferManager {
	return &BufferManager{
		buffers: make(map[string]*Buffer),
	}
}

func (bm *BufferManager) Open(path string) *Buffer {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if buf, ok := bm.buffers[path]; ok {
		bm.active = buf
		return buf
	}

	buf := NewBuffer(path)
	bm.buffers[path] = buf
	bm.active = buf
	return buf
}

func (bm *BufferManager) Active() *Buffer {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.active
}

func (bm *BufferManager) SaveActive() {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	if bm.active != nil {
		bm.active.Save()
	}
}

func (buf *Buffer) DeleteAtCursor(cursorX, cursorY int, selStartY, selEndY int, selecting bool) {
	if selecting {
		// Delete all lines between selStartY and selEndY
		start, end := selStartY, selEndY
		if start > end {
			start, end = end, start
		}
		buf.Content = append(buf.Content[:start], buf.Content[end+1:]...)
		buf.CursorY = start
		buf.CursorX = 0
	} else {
		// Delete single character at cursor
		if cursorY < len(buf.Content) && cursorX < len(buf.Content[cursorY]) {
			line := buf.Content[cursorY]
			buf.Content[cursorY] = append(line[:cursorX], line[cursorX+1:]...)
		} else if cursorY < len(buf.Content)-1 {
			// Join next line if at end
			buf.Content[cursorY] = append(buf.Content[cursorY], buf.Content[cursorY+1]...)
			buf.Content = append(buf.Content[:cursorY+1], buf.Content[cursorY+2:]...)
		}
	}
}

// DeleteSelection removes text from selStart -> selEnd (multi-line capable)
func (buf *Buffer) DeleteSelection(selStartX, selStartY, selEndX, selEndY int) {
	// Ensure selection coordinates are ordered
	if selStartY > selEndY || (selStartY == selEndY && selStartX > selEndX) {
		selStartX, selEndX = selEndX, selStartX
		selStartY, selEndY = selEndY, selStartY
	}

	if selStartY == selEndY {
		// Single-line selection
		line := buf.Content[selStartY]
		buf.Content[selStartY] = append(line[:selStartX], line[selEndX:]...)
	} else {
		// Multi-line selection
		startLine := buf.Content[selStartY][:selStartX]
		endLine := buf.Content[selEndY][selEndX:]
		buf.Content[selStartY] = append(startLine, endLine...)

		// Remove intermediate lines
		buf.Content = append(buf.Content[:selStartY+1], buf.Content[selEndY+1:]...)
	}

	// Move cursor to start of selection
	buf.CursorY = selStartY
	buf.CursorX = selStartX
}

// Buffer methods
// CopySelection copies the selected text to clipboard (multi-line)
func (buf *Buffer) CopySelection(selStartX, selStartY, selEndX, selEndY int) []rune {
	startY, endY := selStartY, selEndY
	startX, endX := selStartX, selEndX
	if startY > endY || (startY == endY && startX > endX) {
		startY, endY = endY, startY
		startX, endX = endX, startX
	}

	var copied []rune
	for y := startY; y <= endY; y++ {
		line := buf.Content[y]
		if y == startY && y == endY {
			copied = append(copied, line[startX:endX]...)
		} else if y == startY {
			copied = append(copied, line[startX:]...)
			copied = append(copied, '\n')
		} else if y == endY {
			copied = append(copied, line[:endX]...)
		} else {
			copied = append(copied, line...)
			copied = append(copied, '\n')
		}
	}
	return copied
}

// CutSelection removes the selection and returns the text
func (buf *Buffer) CutSelection(selStartX, selStartY, selEndX, selEndY int) []rune {
	text := buf.CopySelection(selStartX, selStartY, selEndX, selEndY)

	startY, endY := selStartY, selEndY
	startX, endX := selStartX, selEndX
	if startY > endY || (startY == endY && startX > endX) {
		startY, endY = endY, startY
		startX, endX = endX, startX
	}

	if startY == endY {
		buf.Content[startY] = append(buf.Content[startY][:startX], buf.Content[startY][endX:]...)
	} else {
		// first line
		buf.Content[startY] = buf.Content[startY][:startX]
		// last line
		buf.Content[endY] = buf.Content[endY][endX:]
		// merge first and last line
		buf.Content[startY] = append(buf.Content[startY], buf.Content[endY]...)
		// remove intermediate lines
		buf.Content = append(buf.Content[:startY+1], buf.Content[endY+1:]...)
	}

	// Remove blank line if selection covered a full empty line
	if len(buf.Content) > 0 && len(buf.Content[startY]) == 0 {
		buf.Content = append(buf.Content[:startY], buf.Content[startY+1:]...)
	}

	buf.CursorY = startY
	buf.CursorX = startX

	return text
}

// PasteClipboard inserts text at cursor
func (buf *Buffer) PasteClipboard(text []rune) {
	if buf.CursorY >= len(buf.Content) {
		buf.Content = append(buf.Content, []rune{})
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

	// Insert lines at cursor
	line := buf.Content[buf.CursorY]
	before := line[:buf.CursorX]
	after := line[buf.CursorX:]

	buf.Content[buf.CursorY] = append(before, lines[0]...)
	if len(lines) > 1 {
		buf.Content = append(buf.Content[:buf.CursorY+1], append(lines[1:], buf.Content[buf.CursorY+1:]...)...)
	}
	// Append rest of original line to last pasted line
	lastLineIdx := buf.CursorY + len(lines) - 1
	buf.Content[lastLineIdx] = append(buf.Content[lastLineIdx], after...)

	buf.CursorY = lastLineIdx
	buf.CursorX = len(lines[len(lines)-1])
}

func (b *Buffer) ParentFolder() string {
	if b.File == "" {
		return "."
	}
	return filepath.Dir(b.File)
}
