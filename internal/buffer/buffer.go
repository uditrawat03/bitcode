package buffer

import (
	"bufio"
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	lsp "github.com/uditrawat03/bitcode/internal/lsp_client"
)

type Buffer struct {
	Content    [][]rune
	File       string
	CursorX    int
	CursorY    int
	mu         sync.RWMutex
	lsp        *lsp.Client
	debounce   *time.Timer
	debounceMu sync.Mutex

	Diagnostics []lsp.Diagnostic
	HoverInfo   string
	Completions []lsp.CompletionItem
	CodeActions []lsp.CodeAction

	undoStack  []snapshot
	redoStack  []snapshot
	maxHistory int
}

type snapshot struct {
	Content [][]rune
	CursorX int
	CursorY int
}

// ---------------- Constructors -----------------

func NewBuffer(path string, lspServer *lsp.Client) *Buffer {
	buf := &Buffer{File: path, lsp: lspServer}
	if path != "" {
		buf.Load()
	} else {
		buf.Content = [][]rune{{}}
	}

	if buf.lsp != nil && path != "" {
		lang := detectLanguage(path)
		buf.lsp.SendDidOpen(path, buf.ContentAsString(), lang)
	}

	buf.mu.Lock()
	buf.pushUndoLocked() // store initial state
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

	if b.lsp != nil {
		b.lsp.SendDidSave(b.File)
	}
}

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
	if b.lsp == nil || b.File == "" {
		return
	}

	b.debounceMu.Lock()
	defer b.debounceMu.Unlock()

	if b.debounce != nil {
		b.debounce.Stop()
	}

	b.debounce = time.AfterFunc(50*time.Millisecond, func() {
		b.mu.RLock()
		text := b.contentAsStringLocked()
		b.mu.RUnlock()
		b.lsp.SendDidChange(b.File, text)
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

// ---------------- LSP Features -----------------

func (b *Buffer) UpdateDiagnostics(diags []lsp.Diagnostic) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Diagnostics = diags
}

func (b *Buffer) RequestHover(ctx context.Context, pos lsp.Position) error {
	if b.lsp == nil {
		return nil
	}
	hover, err := b.lsp.Hover(ctx, b.File, pos)
	if err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.HoverInfo = hover
	return nil
}

func (b *Buffer) RequestCompletions(ctx context.Context, pos lsp.Position) error {
	if b.lsp == nil {
		return nil
	}
	items, err := b.lsp.Completion(ctx, b.File, pos)
	if err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Completions = items
	return nil
}

func (b *Buffer) RequestCodeActions(ctx context.Context, rng lsp.Range) error {
	if b.lsp == nil {
		return nil
	}
	actions, err := b.lsp.CodeAction(ctx, b.File, rng)
	if err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.CodeActions = actions
	return nil
}

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

func (b *Buffer) DeleteSelection(selStartX, selStartY, selEndX, selEndY int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.pushUndoLocked()

	sx, sy, ex, ey := normalizeSelection(selStartX, selStartY, selEndX, selEndY)
	if sy == ey {
		line := b.Content[sy]
		b.Content[sy] = append(line[:sx], line[ex:]...)
	} else {
		startLine := b.Content[sy][:sx]
		endLine := b.Content[ey][ex:]
		b.Content[sy] = append(startLine, endLine...)
		b.Content = append(b.Content[:sy+1], b.Content[ey+1:]...)
	}
	b.CursorY, b.CursorX = sy, sx
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

// ---------------- Undo / Redo -----------------

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

// ---------------- BufferManager -----------------

type BufferManager struct {
	buffers map[string]*Buffer
	active  *Buffer
	mu      sync.RWMutex
	lsp     *lsp.Client
}

func NewBufferManager(lspServer *lsp.Client) *BufferManager {
	bm := &BufferManager{buffers: make(map[string]*Buffer), lsp: lspServer}

	if lspServer != nil {
		// Hook diagnostics into buffers
		lspServer.OnDiagnostics(func(uri string, diags []lsp.Diagnostic) {
			bm.mu.RLock()
			defer bm.mu.RUnlock()
			if buf, ok := bm.buffers[uri]; ok {
				buf.UpdateDiagnostics(diags)
			}
		})
	}

	return bm
}

func (bm *BufferManager) Open(path string) *Buffer {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	if buf, ok := bm.buffers[path]; ok {
		bm.active = buf
		return buf
	}
	buf := NewBuffer(path, bm.lsp)
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

func (bm *BufferManager) Close(path string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	if buf, ok := bm.buffers[path]; ok {
		delete(bm.buffers, path)
		if bm.lsp != nil {
			bm.lsp.SendDidClose(path)
		}
		if bm.active == buf {
			bm.active = nil
		}
	}
}

func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "go"
	case ".ts":
		return "typescript"
	case ".tsx":
		return "typescriptreact"
	case ".js":
		return "javascript"
	case ".jsx":
		return "javascriptreact"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp", ".cc", ".cxx":
		return "cpp"
	case ".rs":
		return "rust"
	case ".cs":
		return "csharp"
	case ".php":
		return "php"
	case ".sh", ".bash", ".zsh":
		return "shellscript"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".md":
		return "markdown"
	default:
		return "plaintext"
	}
}
