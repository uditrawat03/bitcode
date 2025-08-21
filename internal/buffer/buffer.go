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
	log.Println("File saved.")
}

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

// Open a file into a buffer (or reuse existing one)
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
