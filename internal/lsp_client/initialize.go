package lsp

// --------------------
// Initialize Request
// --------------------
type InitializeParams struct {
	ProcessID    int                `json:"processId,omitempty"`
	ClientInfo   ClientInfo         `json:"clientInfo,omitempty"`
	RootURI      string             `json:"rootUri,omitempty"`
	Capabilities ClientCapabilities `json:"capabilities"`
}

type ClientCapabilities struct {
	TextDocument TextDocumentClientCapabilities `json:"textDocument,omitempty"`
	Workspace    WorkspaceClientCapabilities    `json:"workspace,omitempty"`
}

type TextDocumentClientCapabilities struct {
	Synchronization *TextDocumentSyncClientCapabilities `json:"synchronization,omitempty"`
	Completion      *CompletionClientCapabilities       `json:"completion,omitempty"`
	Hover           *struct{}                           `json:"hover,omitempty"`
	Definition      *struct{}                           `json:"definition,omitempty"`
}

type TextDocumentSyncClientCapabilities struct {
	DidSave bool `json:"didSave,omitempty"`
}

type CompletionClientCapabilities struct {
	CompletionItem *CompletionItemCapabilities `json:"completionItem,omitempty"`
}

type CompletionItemCapabilities struct {
	SnippetSupport bool `json:"snippetSupport,omitempty"`
}

type WorkspaceClientCapabilities struct {
	ApplyEdit bool `json:"applyEdit,omitempty"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// --------------------
// Initialize Response
// --------------------
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo,omitempty"`
}

// --- Server Capabilities (server responds with this) ---
type ServerCapabilities struct {
	HoverProvider           bool                  `json:"hoverProvider,omitempty"`
	DefinitionProvider      bool                  `json:"definitionProvider,omitempty"`
	CompletionProvider      *CompletionOptions    `json:"completionProvider,omitempty"`
	TextDocumentSync        *TextDocumentSyncKind `json:"textDocumentSync,omitempty"`
	CodeActionProvider      bool                  `json:"codeActionProvider,omitempty"`
	WorkspaceSymbolProvider bool                  `json:"workspaceSymbolProvider,omitempty"`
	// ... add more as needed from LSP spec
}

type CompletionOptions struct {
	ResolveProvider   bool     `json:"resolveProvider,omitempty"`
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

type TextDocumentSyncKind int

const (
	TextDocumentSyncNone        TextDocumentSyncKind = 0
	TextDocumentSyncFull        TextDocumentSyncKind = 1
	TextDocumentSyncIncremental TextDocumentSyncKind = 2
)

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}
