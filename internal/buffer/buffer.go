package buffer

import (
	"bufio"
	"log"
	"os"
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
