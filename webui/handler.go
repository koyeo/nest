package webui

// EventHandler is the interface alias that runner.StepEventHandler implements.
// We define it here to avoid import cycles.
type EventHandler interface {
	OnStepStart(index int, name string)
	OnStepDone(index int, err error)
	OnOutput(content string)
	OnTaskDone(err error)
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

	// Signal the server to shut down after a delay
	go func() {
		// Keep server alive so browser can render final state
		// Server will close when the process exits
		select {}
	}()
}
