package webui

import (
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
func RunWithUI(taskName string, stepNames []string, stepDetails []StepDetail, projectName, projectPath string, execFn func(h EventHandler)) {
	srv := &uiServer{
		taskName:    taskName,
		stepNames:   stepNames,
		stepDetails: stepDetails,
		projectName: projectName,
		projectPath: projectPath,
		execFn:      execFn,
		done:        make(chan struct{}),
		clientReady: make(chan struct{}),
	}
	srv.run()
}

type uiServer struct {
	taskName    string
	stepNames   []string
	stepDetails []StepDetail
	projectName string
	projectPath string
	execFn      func(h EventHandler)
	done        chan struct{}
	clientReady chan struct{}
	readyOnce   sync.Once

	mu      sync.Mutex
	clients []*websocket.Conn
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

	// Start task execution in background (waits for client to connect first)
	go s.startTaskExecution()

	// Open UI — this BLOCKS on macOS (webview needs main thread)
	// When the window closes, it returns and we shut down
	if runtime.GOOS == "darwin" {
		openWebview(url, s.taskName)
	} else {
		openBrowser(url)
		<-s.done // Wait for task to complete
	}

	_ = httpServer.Close()
}

func (s *uiServer) startTaskExecution() {
	// Wait for WebSocket client to connect
	select {
	case <-s.clientReady:
	case <-time.After(10 * time.Second):
		fmt.Fprintf(os.Stderr, "webui: no client connected, aborting\n")
		close(s.done)
		return
	}

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

	// Execute the task
	s.execFn(handler)

	// Small delay to ensure final output is flushed
	time.Sleep(200 * time.Millisecond)

	// Restore stdout/stderr
	stdoutW.Close()
	stderrW.Close()
	os.Stdout = origStdout
	os.Stderr = origStderr

	// Signal done (for non-webview platforms)
	select {
	case <-s.done:
	default:
		close(s.done)
	}
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
	s.mu.Unlock()

	// Signal that a client has connected
	s.readyOnce.Do(func() { close(s.clientReady) })

	// Send init message with task info
	s.sendToConn(conn, map[string]interface{}{
		"type":         "init",
		"task_name":    s.taskName,
		"steps":        s.stepNames,
		"step_details": s.stepDetails,
		"project_name": s.projectName,
		"project_path": s.projectPath,
	})

	// Keep reading to detect close
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
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

func (s *uiServer) sendToConn(conn *websocket.Conn, msg interface{}) {
	data, _ := json.Marshal(msg)
	_ = conn.WriteMessage(websocket.TextMessage, data)
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
