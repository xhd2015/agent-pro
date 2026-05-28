package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/xhd2015/agent-traces/agent/opencode"
)

type message struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionID,omitempty"`
	Prompt    string `json:"prompt,omitempty"`
	Model     string `json:"model,omitempty"`
	Agent     string `json:"agent,omitempty"`
	Command   string `json:"command,omitempty"`
	Variant   string `json:"variant,omitempty"`
}

type Server struct {
	agent *opencode.OpencodeAgent
}

func NewServer(agent *opencode.OpencodeAgent) *Server {
	return &Server{agent: agent}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("websocket upgrade: %v", err), http.StatusBadRequest)
		return
	}
	defer c.Close(websocket.StatusNormalClosure, "done")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var mu sync.Mutex
	writeJSON := func(v any) error {
		mu.Lock()
		defer mu.Unlock()
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		return wsjson.Write(ctx, c, v)
	}

	for {
		_, msgBytes, err := c.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				return
			}
			return
		}

		var msg message
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			writeJSON(map[string]string{"type": "error", "error": fmt.Sprintf("invalid message: %v", err)})
			continue
		}

		switch msg.Type {
		case "start_session":
			opts := &opencode.SessionOpts{}
			if msg.Model != "" {
				opts.Model = msg.Model
			}
			if msg.Agent != "" {
				opts.Agent = msg.Agent
			}
			if msg.Command != "" {
				opts.Command = msg.Command
			}
			if msg.Variant != "" {
				opts.Variant = msg.Variant
			}

			onEvent := func(event opencode.StreamEvent) {
				writeJSON(event)
			}

			sessionID, runErr := s.agent.StartSession(ctx, msg.Prompt, opts, onEvent)
			if runErr != nil {
				writeJSON(map[string]string{"type": "error", "error": runErr.Error()})
				return
			}
			writeJSON(map[string]string{"type": "done", "sessionID": sessionID})

		case "resume_session":
			if msg.SessionID == "" {
				writeJSON(map[string]string{"type": "error", "error": "sessionID required"})
				continue
			}

			opts := &opencode.SessionOpts{}
			if msg.Model != "" {
				opts.Model = msg.Model
			}
			if msg.Agent != "" {
				opts.Agent = msg.Agent
			}
			if msg.Command != "" {
				opts.Command = msg.Command
			}
			if msg.Variant != "" {
				opts.Variant = msg.Variant
			}

			onEvent := func(event opencode.StreamEvent) {
				writeJSON(event)
			}

			_, runErr := s.agent.ResumeSession(ctx, msg.SessionID, msg.Prompt, opts, onEvent)
			if runErr != nil {
				writeJSON(map[string]string{"type": "error", "error": runErr.Error()})
				return
			}
			writeJSON(map[string]string{"type": "done"})

		default:
			writeJSON(map[string]string{"type": "error", "error": fmt.Sprintf("unknown message type: %s", msg.Type)})
		}
	}
}
