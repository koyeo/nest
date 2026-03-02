package webui

import "encoding/json"

// EventHandler is the interface alias that runner.StepEventHandler implements.
// We define it here to avoid import cycles.
type EventHandler interface {
	OnStepStart(index int, name string)
	OnStepDone(index int, err error)
	OnOutput(content string)
	OnTaskDone(err error)
	// Prompt shows a message to the user and returns their text input.
	Prompt(message string) string
}

// wsHandler implements EventHandler by broadcasting events via WebSocket.
type wsHandler struct {
	server *uiServer
}

func (h *wsHandler) OnStepStart(index int, name string) {
	h.server.broadcast(map[string]interface{}{
		"type":  "step_start",
		"index": index,
		"name":  name,
	})
}

func (h *wsHandler) OnStepDone(index int, err error) {
	msg := map[string]interface{}{
		"type":  "step_done",
		"index": index,
	}
	if err != nil {
		msg["error"] = err.Error()
	}
	h.server.broadcast(msg)
}

func (h *wsHandler) OnOutput(content string) {
	h.server.broadcast(map[string]interface{}{
		"type":    "output",
		"content": content,
	})
}

func (h *wsHandler) OnTaskDone(err error) {
	msg := map[string]interface{}{
		"type": "task_done",
	}
	if err != nil {
		msg["error"] = err.Error()
	}
	h.server.broadcast(msg)

	// Keep server alive so browser can render final state
	go func() {
		select {}
	}()
}

// Prompt sends a prompt message to the browser and blocks until the user responds.
func (h *wsHandler) Prompt(message string) string {
	// Create a response channel
	ch := make(chan string, 1)
	h.server.setPromptCh(ch)

	// Broadcast the prompt request
	h.server.broadcast(map[string]interface{}{
		"type":    "prompt",
		"message": message,
	})

	// Block until user responds
	response := <-ch
	return response
}

// handlePromptResponse is called when a prompt_response message arrives from the client.
func (s *uiServer) handlePromptResponse(raw []byte) {
	var msg struct {
		Type     string `json:"type"`
		Response string `json:"response"`
	}
	if json.Unmarshal(raw, &msg) == nil && msg.Type == "prompt_response" {
		s.execMu.Lock()
		ch := s.promptCh
		s.execMu.Unlock()
		if ch != nil {
			ch <- msg.Response
		}
	}
}

func (s *uiServer) setPromptCh(ch chan string) {
	s.execMu.Lock()
	s.promptCh = ch
	s.execMu.Unlock()
}
