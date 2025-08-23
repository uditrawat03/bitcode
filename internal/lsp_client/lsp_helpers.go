package lsp

import (
	"os"
	"path/filepath"
	"strings"
)

func (s *Client) Initialize(rootURI string) (chan Message, error) {
	params := InitializeParams{
		ProcessID: os.Getpid(),
		ClientInfo: ClientInfo{
			Name:    "bitcode",
			Version: "0.0.0",
		},
		RootURI: rootURI,
		Capabilities: ClientCapabilities{
			TextDocument: TextDocumentClientCapabilities{
				Synchronization: &TextDocumentSyncClientCapabilities{
					DidSave: true,
				},
				Completion: &CompletionClientCapabilities{
					CompletionItem: &CompletionItemCapabilities{
						SnippetSupport: true,
					},
				},
				Hover:      &struct{}{},
				Definition: &struct{}{},
			},
			Workspace: WorkspaceClientCapabilities{
				ApplyEdit: true,
			},
		},
	}

	return s.SendRequest(s.ctx, "initialize", params)
}

func (s *Client) Initialized() {
	s.SendNotification("initialized", struct{}{})
}

func (s *Client) SendDidOpen(filePath, content, language string) {
	uri := pathToURI(filePath)

	s.versionMu.Lock()
	s.fileVer[uri] = 1
	version := s.fileVer[uri]
	s.versionMu.Unlock()

	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        uri,
			LanguageID: language,
			Version:    version,
			Text:       content,
		},
	}

	s.logger.Println("[SendDidOpen]", uri, "v", version)

	s.SendNotification("textDocument/didOpen", params)
}

func (s *Client) SendDidChange(filePath, content string) {
	uri := pathToURI(filePath)

	s.versionMu.Lock()
	s.fileVer[uri]++
	version := s.fileVer[uri]
	s.versionMu.Unlock()

	params := DidChangeTextDocumentParams{
		TextDocument: VersionedTextDocumentIdentifier{
			URI:     uri,
			Version: version,
		},
		ContentChanges: []TextDocumentContentChangeEvent{
			{Text: content},
		},
	}

	s.SendNotification("textDocument/didChange", params)
}

func (s *Client) SendDidSave(filePath string) {
	uri := pathToURI(filePath)

	params := map[string]interface{}{
		"textDocument": map[string]string{"uri": uri},
	}
	s.SendNotification("textDocument/didSave", params)
}

func (s *Client) SendDidClose(filePath string) {
	uri := pathToURI(filePath)

	params := map[string]interface{}{
		"textDocument": map[string]string{"uri": uri},
	}
	s.SendNotification("textDocument/didClose", params)
}

func pathToURI(path string) string {
	abs, _ := filepath.Abs(TrimUri(path))
	return "file://" + filepath.ToSlash(abs)
}

func TrimUri(uri string) string {
	return strings.TrimPrefix(uri, "file://")
}
