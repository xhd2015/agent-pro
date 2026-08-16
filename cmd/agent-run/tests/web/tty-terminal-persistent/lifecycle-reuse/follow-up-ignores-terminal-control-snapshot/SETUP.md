# Scenario

**Bug**: a reused live terminal follow-up captures terminal control/snapshot
metadata as the assistant response.

```
finished tty chat with live mapped terminal
  -> follow-up sender writes prompt to existing PTY
  -> terminal snapshot contains prompt plus session/control JSON but no answer
  -> web chat must not append that snapshot JSON as assistant response
```

## Preconditions

- Session metadata maps the web chat to `terminal_session_id:"session-1"`.
- The fake terminal websocket accepts input but never emits a real assistant
  answer.
- Reattaching for a snapshot returns terminal prompt text followed by
  session/control-looking JSON.

## Steps

1. Write a finished mapped `codex-tty` session fixture.
2. Write a live registry entry pointing at a ptywrap-like websocket server.
3. POST a follow-up message through the web API.
4. Poll the session events.
5. Assert no assistant message contains `session_id`, `terminal_session_id`,
   prompt echo, or terminal prompt text.

```go
import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "followup"
	req.Runner = "codex-tty"
	req.ChatSessionID = "web_control_snapshot_followup"
	req.TerminalSessionID = "session-1"
	req.Prompt = "one word of France capital"
	req.FollowUpPrompt = "What did I say"
	writeMappedSessionFixture(t, req)
	listenAddr := startNoAnswerControlSnapshotPtywrap(t, req)
	writeTTYRegistryFixture(t, req, req.TerminalSessionID, listenAddr)
	return nil
}

func startNoAnswerControlSnapshotPtywrap(t *testing.T, req *Request) string {
	t.Helper()
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/terminal", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		requestedID := r.URL.Query().Get("session_id")
		if requestedID == "" {
			requestedID = req.TerminalSessionID
		}
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"session_id","session_id":"`+requestedID+`"}`))
		snapshot := "CODEX_TTY_BANNER\r\nCodex › " + req.FollowUpPrompt +
			"\r\n" + `{"session_id":"` + req.TerminalSessionID + `","terminal_session_id":"` + req.TerminalSessionID + `","status":"running"}` + "\r\n"
		_ = conn.WriteMessage(websocket.BinaryMessage, []byte(snapshot))
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return strings.TrimPrefix(server.URL, "http://")
}

func sessionDetailBodyEventually(t *testing.T, req *Request, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var last string
	for time.Now().Before(deadline) {
		status, body := doHTTP(t, "GET", req.WebBaseURL+"/api/agent-run/sessions/"+req.ChatSessionID, req.WebToken, "", "")
		if status != http.StatusOK {
			t.Fatalf("session detail status=%d body=%s", status, body)
		}
		last = body
		if strings.Contains(body, "session_id") || strings.Contains(body, req.FollowUpPrompt) {
			return body
		}
		time.Sleep(100 * time.Millisecond)
	}
	return last
}
```
