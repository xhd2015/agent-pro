package subagent

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/xhd2015/agent-pro/agent/event/logging"
)

type sessionLogWriter struct {
	mu         sync.Mutex
	sessionID  string
	eventsFile logging.Logger
}

func (w *sessionLogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.eventsFile != nil {
		_ = w.eventsFile.Append(p)
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
