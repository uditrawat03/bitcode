package buffer

import (
	"context"
	"path/filepath"
	"strings"
	"sync"

	lsp "github.com/uditrawat03/bitcode/internal/lsp_client"
)

// ---------------- BufferManager -----------------

type BufferManager struct {
	ctx     context.Context
	buffers map[string]*Buffer
	active  *Buffer
	mu      sync.RWMutex
	lsp     *lsp.Client
}

func NewBufferManager(ctx context.Context, lspServer *lsp.Client) *BufferManager {
	bm := &BufferManager{buffers: make(map[string]*Buffer), lsp: lspServer, ctx: ctx}

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

	if bm.lsp != nil {
		bm.lsp.OnDiagnostics(func(uri string, diags []lsp.Diagnostic) {
			bm.mu.RLock()
			buf, ok := bm.buffers[TrimUri(uri)]
			bm.mu.RUnlock()
			if !ok {
				return
			}

			buf.UpdateDiagnostics(diags)

			for _, d := range diags {
				go buf.RequestCodeActions(bm.ctx, d.Range)
			}
		})

		bm.lsp.OnCodeAction(func(uri string, actions []lsp.CodeAction) {
			bm.mu.RLock()
			defer bm.mu.RUnlock()
			if buf, ok := bm.buffers[TrimUri(uri)]; ok {
				buf.mu.Lock()
				buf.CodeActions = actions
				buf.mu.Unlock()
			}
		})
	}

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
			bm.lsp.SendDidClose(bm.active.URI())
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
