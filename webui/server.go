package webui

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed page.html
var pageFS embed.FS

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// StepDetail mirrors runner.StepDetail for JSON serialization.
type StepDetail struct {
	Name    string `json:"name"`
	Depth   int    `json:"depth"`
	IsGroup bool   `json:"is_group"`
}

// RunWithUI starts the webui server, opens a window/browser, executes the task,
// and streams events in real-time via WebSocket.
// On macOS, this function blocks on the webview window (main thread requirement).
func RunWithUI(taskName string, stepNames []string, stepDetails []StepDetail, projectName, projectPath string, execFn func(h EventHandler, ctx context.Context)) {
	srv := &uiServer{
		taskName:    taskName,
		stepNames:   stepNames,
		stepDetails: stepDetails,
		projectName: projectName,
		projectPath: projectPath,
		execFn:      execFn,
		done:        make(chan struct{}),
		clientReady: make(chan struct{}),
		actionCh:    make(chan string, 10),
	}
	srv.run()
}

type uiServer struct {
	taskName    string
	stepNames   []string
	stepDetails []StepDetail
	projectName string
	projectPath string
	execFn      func(h EventHandler, ctx context.Context)
	done        chan struct{}
	clientReady chan struct{}
	readyOnce   sync.Once
	actionCh    chan string // "stop", "rerun"

	mu      sync.Mutex
	clients []*websocket.Conn

	// Execution state
	execMu  sync.Mutex
	running bool
	cancel  context.CancelFunc // cancel current execution context
}

func (s *uiServer) run() {
	// Find a free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start webui server: %v\n", err)
		return
	}
	addr := listener.Addr().String()
	url := fmt.Sprintf("http://%s", addr)

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handlePage)
	mux.HandleFunc("/ws", s.handleWS)

	httpServer := &http.Server{Handler: mux}

	// Start HTTP server in background
	go func() {
		_ = httpServer.Serve(listener)
	}()

	// Start action loop: handles run/rerun/stop commands
	go s.actionLoop()

	// Open UI — this BLOCKS on macOS (webview needs main thread)
	if runtime.GOOS == "darwin" {
		openWebview(url, s.taskName)
	} else {
		openBrowser(url)
		<-s.done
	}

	_ = httpServer.Close()
}

// actionLoop handles start/stop/rerun lifecycle.
func (s *uiServer) actionLoop() {
	// Wait for WebSocket client to connect before first run
	select {
	case <-s.clientReady:
	case <-time.After(10 * time.Second):
		fmt.Fprintf(os.Stderr, "webui: no client connected, aborting\n")
		close(s.done)
		return
	}

	// Auto-run on connect
	s.runTask()

	// Listen for actions from WebSocket
	for action := range s.actionCh {
		switch action {
		case "stop":
			s.stopTask()
		case "rerun":
			s.stopTask()
			// Small delay to let pipes flush
			time.Sleep(300 * time.Millisecond)
			// Broadcast reset to clear UI
			s.broadcast(map[string]interface{}{
				"type": "reset",
			})
			s.runTask()
		}
	}
}

func (s *uiServer) runTask() {
	s.execMu.Lock()
	if s.running {
		s.execMu.Unlock()
		return
	}
	s.running = true
	s.execMu.Unlock()

	// Broadcast state
	s.broadcast(map[string]interface{}{
		"type":  "state",
		"state": "running",
	})

	go s.executeTask()
}

func (s *uiServer) stopTask() {
	s.execMu.Lock()
	if !s.running {
		s.execMu.Unlock()
		return
	}
	// Cancel context → sends SIGKILL to local subprocesses, closes SSH sessions
	if s.cancel != nil {
		s.cancel()
	}
	s.execMu.Unlock()

	// Wait for execution to finish
	for {
		s.execMu.Lock()
		r := s.running
		s.execMu.Unlock()
		if !r {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func (s *uiServer) executeTask() {
	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	s.execMu.Lock()
	s.cancel = cancel
	s.execMu.Unlock()

	// Intercept os.Stdout and os.Stderr to capture ALL output
	origStdout := os.Stdout
	origStderr := os.Stderr
	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()
	os.Stdout = stdoutW
	os.Stderr = stderrW

	handler := &wsHandler{server: s}

	// Pump intercepted stdout/stderr to WebSocket
	go s.pumpPipe(stdoutR)
	go s.pumpPipe(stderrR)

	// Execute the task with context
	s.execFn(handler, ctx)

	// Small delay to ensure final output is flushed
	time.Sleep(200 * time.Millisecond)

	// Restore stdout/stderr
	_ = stdoutW.Close()
	_ = stderrW.Close()
	os.Stdout = origStdout
	os.Stderr = origStderr

	s.execMu.Lock()
	s.running = false
	s.cancel = nil
	s.execMu.Unlock()

	// Broadcast state
	s.broadcast(map[string]interface{}{
		"type":  "state",
		"state": "stopped",
	})
}

func (s *uiServer) handlePage(w http.ResponseWriter, r *http.Request) {
	data, _ := pageFS.ReadFile("page.html")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func (s *uiServer) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	s.mu.Lock()
	s.clients = append(s.clients, conn)

	// Send init message UNDER LOCK (before any broadcast can happen)
	initData, _ := json.Marshal(map[string]interface{}{
		"type":         "init",
		"task_name":    s.taskName,
		"steps":        s.stepNames,
		"step_details": s.stepDetails,
		"project_name": s.projectName,
		"project_path": s.projectPath,
	})
	_ = conn.WriteMessage(websocket.TextMessage, initData)
	s.mu.Unlock()

	// Signal AFTER init is sent
	s.readyOnce.Do(func() { close(s.clientReady) })

	// Read incoming messages (actions from client)
	go func() {
		for {
			_, raw, err := conn.ReadMessage()
			if err != nil {
				break
			}
			var msg struct {
				Type   string `json:"type"`
				Action string `json:"action"`
			}
			if json.Unmarshal(raw, &msg) == nil && msg.Type == "action" {
				select {
				case s.actionCh <- msg.Action:
				default:
				}
			}
		}
	}()
}

func (s *uiServer) broadcast(msg interface{}) {
	data, _ := json.Marshal(msg)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, conn := range s.clients {
		_ = conn.WriteMessage(websocket.TextMessage, data)
	}
}

func (s *uiServer) pumpPipe(r *os.File) {
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			s.broadcast(map[string]interface{}{
				"type":    "output",
				"content": string(buf[:n]),
			})
		}
		if err != nil {
			break
		}
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	if cmd != nil {
		_ = cmd.Start()
	}
}
