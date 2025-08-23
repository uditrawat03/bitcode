package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// Server represents a running LSP server
type Client struct {
	ctx        context.Context
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	logger     *log.Logger
	fileVer    map[string]int
	versionMu  sync.Mutex
	outgoing   chan Message
	stopSignal chan struct{}

	nextID          int
	pendingRequests map[int]chan Message
	handlers        map[string]func(context.Context, json.RawMessage)

	mu sync.Mutex
}

// JSON-RPC message structure
type Message struct {
	Jsonrpc string           `json:"jsonrpc"`
	ID      *int             `json:"id,omitempty"`
	Method  string           `json:"method,omitempty"`
	Params  *json.RawMessage `json:"params,omitempty"`
	Result  *json.RawMessage `json:"result,omitempty"`
	Error   *ResponseError   `json:"error,omitempty"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewServer creates a new LSP server
func NewServer(ctx context.Context, logger *log.Logger) *Client {
	return &Client{
		ctx:             ctx,
		logger:          logger,
		fileVer:         make(map[string]int),
		outgoing:        make(chan Message, 100),
		stopSignal:      make(chan struct{}),
		pendingRequests: make(map[int]chan Message),
		handlers:        make(map[string]func(context.Context, json.RawMessage)),
	}
}

// Start launches the LSP server and begins loops
func (s *Client) Start(pathToServer string, args ...string) error {
	s.logger.Println("Starting LSP server...")

	s.cmd = exec.Command(pathToServer, args...)
	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderr, err := s.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}
	stdin, err := s.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	s.stdin = stdin
	s.stdout = stdout

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start LSP server: %w", err)
	}

	s.logger.Printf("LSP server started (PID %d)", s.cmd.Process.Pid)

	go s.readLoop(stdout)
	go s.logStderr(stderr)
	go s.writeLoop()

	return nil
}

// Stop the LSP server
func (s *Client) Stop() {
	close(s.stopSignal)
	if s.stdin != nil {
		_ = s.stdin.Close()
	}
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		_, _ = s.cmd.Process.Wait()
	}
	s.logger.Println("LSP server stopped")
}

// ---------------------- Messaging ----------------------

func (s *Client) SendRequest(ctx context.Context, method string, params interface{}) (chan Message, error) {
	s.mu.Lock()
	s.nextID++
	id := s.nextID
	ch := make(chan Message, 1)
	s.pendingRequests[id] = ch
	s.mu.Unlock()

	msg := Message{
		Jsonrpc: "2.0",
		ID:      &id,
		Method:  method,
	}

	if params != nil {
		raw, _ := json.Marshal(params)
		rm := json.RawMessage(raw)
		msg.Params = &rm
	}

	s.SendMessage(msg)
	return ch, nil
}

func (s *Client) SendNotification(method string, params interface{}) {
	msg := Message{
		Jsonrpc: "2.0",
		Method:  method,
	}

	if params != nil {
		raw, _ := json.Marshal(params)
		rm := json.RawMessage(raw)
		msg.Params = &rm
	}

	s.SendMessage(msg)
}

// Low-level queue
func (s *Client) SendMessage(msg Message) {
	select {
	case s.outgoing <- msg:
	default:
		s.logger.Println("LSP outgoing channel full, dropping message")
	}
}

func (s *Client) writeLoop() {
	for {
		select {
		case msg := <-s.outgoing:
			data, _ := json.Marshal(msg)
			message := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(data), string(data))
			if _, err := io.WriteString(s.stdin, message); err != nil {
				s.logger.Printf("Failed to send LSP message: %v", err)
			}
		case <-s.stopSignal:
			return
		}
	}
}

func (s *Client) readLoop(stdout io.ReadCloser) {
	reader := bufio.NewReader(stdout)
	for {
		select {
		case <-s.stopSignal:
			return
		default:
		}

		header, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				s.logger.Println("LSP server closed stdout")
			}
			return
		}

		if !strings.HasPrefix(header, "Content-Length:") {
			continue
		}

		var contentLength int
		fmt.Sscanf(header, "Content-Length: %d", &contentLength)
		_, _ = reader.ReadString('\n') // skip blank line

		content := make([]byte, contentLength)
		if _, err = io.ReadFull(reader, content); err != nil {
			s.logger.Printf("Error reading JSON-RPC content: %v", err)
			continue
		}

		var msg Message
		if err := json.Unmarshal(content, &msg); err != nil {
			s.logger.Printf("Failed to unmarshal JSON-RPC: %v", err)
			continue
		}

		s.dispatch(msg)
	}
}

func (s *Client) logStderr(stderr io.ReadCloser) {
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		s.logger.Printf("LSP stderr: %s", scanner.Text())
	}
}

// ---------------------- Dispatch ----------------------

func (s *Client) dispatch(msg Message) {
	if msg.ID != nil {
		s.mu.Lock()
		if ch, ok := s.pendingRequests[*msg.ID]; ok {
			ch <- msg
			close(ch)
			delete(s.pendingRequests, *msg.ID)
		}
		s.mu.Unlock()
		return
	}

	if msg.Method != "" && msg.Params != nil {
		if handler, ok := s.handlers[msg.Method]; ok {
			handler(s.ctx, *msg.Params)
		}
	}
}

// RegisterHandler lets IDE components listen for notifications
func (s *Client) RegisterHandler(method string, handler func(context.Context, json.RawMessage)) {
	s.handlers[method] = handler
}

// ---------------------- LSP Helpers ----------------------

func (s *Client) Initialize(rootURI string) (chan Message, error) {
	return s.SendRequest(s.ctx, "initialize", map[string]interface{}{
		"processId": os.Getpid(),
		"clientInfo": map[string]interface{}{
			"name":    "bitcode",
			"version": "0.0.0",
		},
		"rootUri":      rootURI,
		"capabilities": map[string]interface{}{},
	})
}

func (s *Client) Initialized() {
	s.SendNotification("initialized", map[string]interface{}{})
}

func (s *Client) SendDidOpen(filePath, content, language string) {
	s.versionMu.Lock()
	s.fileVer[filePath] = 1
	version := s.fileVer[filePath]
	s.versionMu.Unlock()

	s.SendNotification("textDocument/didOpen", map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":        filePath,
			"languageId": language,
			"version":    version,
			"text":       content,
		},
	})
}

func (s *Client) SendDidChange(filePath, content string) {
	s.versionMu.Lock()
	s.fileVer[filePath]++
	version := s.fileVer[filePath]
	s.versionMu.Unlock()

	s.SendNotification("textDocument/didChange", map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":     filePath,
			"version": version,
		},
		"contentChanges": []map[string]string{
			{"text": content},
		},
	})
}

func (s *Client) SendDidSave(filePath string) {
	s.SendNotification("textDocument/didSave", map[string]interface{}{
		"textDocument": map[string]string{"uri": filePath},
	})
}

func (s *Client) SendDidClose(filePath string) {
	s.SendNotification("textDocument/didClose", map[string]interface{}{
		"textDocument": map[string]string{"uri": filePath},
	})
}
