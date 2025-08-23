package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
		raw, err := json.Marshal(params)
		if err != nil {
			s.logger.Printf("failed to marshal params for %s: %v", method, err)
			return
		}
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

		// ---- 1. Read headers ----
		headers := make(map[string]string)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					s.logger.Println("LSP server closed stdout")
				} else {
					s.logger.Printf("Error reading header: %v", err)
				}
				return
			}

			line = strings.TrimSpace(line)
			if line == "" {
				break // end of headers
			}

			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}

		// ---- 2. Parse content length ----
		contentLengthStr, ok := headers["Content-Length"]
		if !ok {
			s.logger.Println("Missing Content-Length header")
			continue
		}

		var contentLength int
		fmt.Sscanf(contentLengthStr, "%d", &contentLength)

		// ---- 3. Read body ----
		content := make([]byte, contentLength)
		if _, err := io.ReadFull(reader, content); err != nil {
			s.logger.Printf("Error reading JSON-RPC body: %v", err)
			continue
		}

		// ---- 4. Parse JSON ----
		var msg Message
		if err := json.Unmarshal(content, &msg); err != nil {
			s.logger.Printf("Failed to unmarshal JSON-RPC: %v", err)
			continue
		}

		// Debug: log raw JSON instead of struct pointers
		s.logger.Printf("[reading loop] %s", content)

		// ---- 5. Dispatch ----
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
