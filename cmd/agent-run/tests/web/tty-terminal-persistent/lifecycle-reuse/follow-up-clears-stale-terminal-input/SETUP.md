# Scenario

**Bug**: a chat follow-up appends to stale text already present in the
backend terminal input line.

```
finished tty chat with live mapped terminal
  -> terminal prompt already contains stale text "Explain this codebase"
  -> follow-up sender writes "what did I say?"
  -> backend must clear/replace stale input before pressing Enter
  -> web chat must not submit "Explain this codebasewhat did I say?"
```

## Preconditions

- Session metadata maps the web chat to `terminal_session_id:"session-1"`.
- The fake terminal websocket snapshots a Codex prompt line with stale input.
- The fake terminal emits `MALFORMED_SUBMISSION` when the follow-up is appended
  to the stale prompt instead of clearing it first.

## Steps

1. Write a finished mapped `codex-tty` session fixture.
2. Write a live registry entry pointing at a ptywrap-like websocket server whose
   prompt line contains `Explain this codebase`.
3. POST a follow-up message `what did I say?` through the web API.
4. Poll the session events.
5. Assert no assistant message contains the stale prompt text or
   `MALFORMED_SUBMISSION`.

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
	req.ChatSessionID = "web_stale_terminal_input_followup"
	req.TerminalSessionID = "session-1"
	req.Prompt = "one word of France capital"
	req.FollowUpPrompt = "what did I say?"
	writeMappedSessionFixture(t, req)
	listenAddr := startStaleInputPtywrap(t, req, "Explain this codebase")
	writeTTYRegistryFixture(t, req, req.TerminalSessionID, listenAddr)
	return nil
}

func startStaleInputPtywrap(t *testing.T, req *Request, staleInput string) string {
	t.Helper()
	upgrader := websocket.Upgrader{}
	inputSeen := ""
	req.PTYInputSeen = &inputSeen
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
		// The stale prompt text is terminal input state, not assistant output.
		// Do not send it as a frame; the server combines it with incoming input
		// unless the client clears the line first.
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if mt != websocket.BinaryMessage {
				continue
			}
			typed := string(msg)
			inputSeen += typed
			submitted := staleInput + strings.TrimRight(typed, "\r\n")
			if strings.Contains(typed, "\x15") {
				parts := strings.Split(typed, "\x15")
				submitted = strings.TrimRight(parts[len(parts)-1], "\r\n")
			}
			if submitted == req.FollowUpPrompt {
				_ = conn.WriteMessage(websocket.BinaryMessage, []byte("\r\nFOLLOWUP_RESPONSE: received "+submitted+"\r\n"))
				continue
			}
			_ = conn.WriteMessage(websocket.BinaryMessage, []byte("\r\nMALFORMED_SUBMISSION: "+submitted+"\r\n"))
		}
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return strings.TrimPrefix(server.URL, "http://")
}

func staleInputSessionBodyEventually(t *testing.T, req *Request, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var last string
	for time.Now().Before(deadline) {
		status, body := doHTTP(t, "GET", req.WebBaseURL+"/api/agent-run/sessions/"+req.ChatSessionID, req.WebToken, "", "")
		if status != http.StatusOK {
			t.Fatalf("session detail status=%d body=%s", status, body)
		}
		last = body
		if strings.Contains(body, "MALFORMED_SUBMISSION") || strings.Contains(body, "FOLLOWUP_RESPONSE") {
			return body
		}
		time.Sleep(100 * time.Millisecond)
	}
	return last
}
```
