package subagent

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/xhd2015/agent-pro/agent/event/convert"
	"github.com/xhd2015/agent-pro/agent/event/logging"
)

type sessionLogWriter struct {
	mu          sync.Mutex
	sessionID   string
	agentRunner string
	eventsFile  logging.Logger
}

func (w *sessionLogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.eventsFile != nil {
		// Try to convert raw bytes to AgentEvent JSONL
		events, err := convert.ConvertRawLine(p, w.agentRunner)
		if err != nil {
			// Fall back: append raw bytes as-is
			_ = w.eventsFile.Append(p)
		} else {
			for _, event := range events {
				data, err := json.Marshal(event)
				if err != nil {
					continue
				}
				_ = w.eventsFile.Append(append(data, '\n'))
			}
		}
	}

	if w.sessionID == "" {
		line := strings.TrimSpace(string(p))
		if line != "" && line[0] == '{' {
			var event struct {
				SessionID string `json:"sessionID,omitempty"`
			}
			if json.Unmarshal([]byte(line), &event) == nil && event.SessionID != "" {
				w.sessionID = event.SessionID
			}
		}
	}

	return len(p), nil
}

func (w *sessionLogWriter) Close() error {
	if w.eventsFile != nil {
		return w.eventsFile.Close()
	}
	return nil
}
