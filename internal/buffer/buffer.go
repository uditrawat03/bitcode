package buffer

import (
	"bufio"
	"context"
	"log"
	"os"
	"sync"
	"time"

	lsp "github.com/uditrawat03/bitcode/internal/lsp_client"
	"github.com/uditrawat03/bitcode/internal/utils"
)

type Buffer struct {
	ctx        context.Context
	Content    [][]rune
	File       string
	CursorX    int
	CursorY    int
	mu         sync.RWMutex
	LspClient  *lsp.Client
	debounce   *time.Timer
	debounceMu sync.Mutex

	Diagnostics []lsp.Diagnostic
	HoverInfo   string
	Completions []lsp.CompletionItem
	CodeActions []lsp.CodeAction

	undoStack  []snapshot
	redoStack  []snapshot
	maxHistory int

	logger *log.Logger
}

type snapshot struct {
	Content [][]rune
	CursorX int
	CursorY int
}

func NewBuffer(ctx context.Context, lspServer *lsp.Client, path string) *Buffer {
	buf := &Buffer{File: path, LspClient: lspServer, ctx: ctx}

	logger, _ := utils.GetLogger("./log/buffer.log")

	buf.logger = logger

	if path != "" {
		buf.Load()
	} else {
		buf.Content = [][]rune{{}}
	}

	if buf.LspClient != nil && path != "" {
		lang := detectLanguage(path)
		buf.LspClient.SendDidOpen(buf.URI(), buf.ContentAsString(), lang)
	}

	buf.mu.Lock()
	buf.pushUndoLocked()
	buf.mu.Unlock()

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

	if b.LspClient != nil {
		b.LspClient.SendDidSave(b.URI())
	}
}
