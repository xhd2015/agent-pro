---
label: e2e
---

## Expected

- The follow-up POST is accepted.
- Terminal snapshot/control metadata is not appended as an assistant chat event.
- Without a real assistant answer, the turn is not marked `done` and the
  session remains `running`.
- A subsequent follow-up POST is still accepted, proving the persistent TTY chat
  remains usable.
- In particular, assistant events must not contain `session_id`,
  `terminal_session_id`, `CODEX_TTY_BANNER`, `Codex ›`, `[Terminal exited]`, or
  the echoed follow-up prompt.

## Exit Code

- Test process exits non-zero while the live-terminal follow-up path converts a
  terminal snapshot or control frame into an assistant response, appends `done`
  without a real answer, or stops accepting follow-ups.

```go
import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.FollowUpStatus != http.StatusAccepted {
		t.Fatalf("follow-up status=%d body=%s", resp.FollowUpStatus, resp.FollowUpBody)
	}

	deadline := time.Now().Add(5 * time.Second)
	var last string
	for time.Now().Before(deadline) {
		status, body := doHTTP(t, "GET", req.WebBaseURL+"/api/agent-run/sessions/"+req.ChatSessionID, req.WebToken, "", "")
		if status != http.StatusOK {
			t.Fatalf("session detail status=%d body=%s", status, body)
		}
		last = body
		time.Sleep(100 * time.Millisecond)
	}
	var violations []string
	if bad := malformedAssistantSnapshot(t, req, last); bad != "" {
		violations = append(violations, "terminal snapshot/control text stored as assistant response: "+bad)
	}
	if hasDoneAfterFollowUp(t, req, last) {
		violations = append(violations, "done event appended after follow-up even though no real assistant answer exists")
	}
	if status := sessionStatus(t, last); status != "running" {
		violations = append(violations, fmt.Sprintf("session status=%q, want running", status))
	}
	secondStatus, secondBody := doHTTP(t, "POST", req.WebBaseURL+"/api/agent-run/sessions/"+req.ChatSessionID+"/messages", req.WebToken, "application/json", `{"text":"Are you still there?"}`)
	if secondStatus != http.StatusAccepted {
		violations = append(violations, fmt.Sprintf("second follow-up status=%d body=%s", secondStatus, secondBody))
	}
	if len(violations) > 0 {
		t.Fatalf("live terminal follow-up should ignore control snapshots and keep session reusable:\n- %s\nbody=%s", strings.Join(violations, "\n- "), last)
	}
}

func malformedAssistantSnapshot(t *testing.T, req *Request, body string) string {
	t.Helper()
	var parsed struct {
		Events []struct {
			Type string `json:"type"`
			Role string `json:"role"`
			Text string `json:"text"`
		} `json:"events"`
	}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("invalid session detail JSON: %v\n%s", err, body)
	}
	for _, ev := range parsed.Events {
		if ev.Type != "message" || ev.Role != "assistant" {
			continue
		}
		text := ev.Text
		if strings.Contains(text, "session_id") ||
			strings.Contains(text, "terminal_session_id") ||
			strings.Contains(text, "CODEX_TTY_BANNER") ||
			strings.Contains(text, "Codex ›") ||
			strings.Contains(text, "[Terminal exited]") ||
			strings.Contains(text, req.FollowUpPrompt) {
			return text
		}
	}
	return ""
}

func hasDoneAfterFollowUp(t *testing.T, req *Request, body string) bool {
	t.Helper()
	var parsed struct {
		Events []struct {
			Type string `json:"type"`
			Role string `json:"role"`
			Text string `json:"text"`
		} `json:"events"`
	}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("invalid session detail JSON: %v\n%s", err, body)
	}
	seenFollowUp := false
	for _, ev := range parsed.Events {
		if ev.Type == "message" && ev.Role == "user" && ev.Text == req.FollowUpPrompt {
			seenFollowUp = true
			continue
		}
		if seenFollowUp && ev.Type == "done" {
			return true
		}
	}
	return false
}

func sessionStatus(t *testing.T, body string) string {
	t.Helper()
	var parsed struct {
		Session struct {
			Status string `json:"status"`
		} `json:"session"`
	}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("invalid session detail JSON: %v\n%s", err, body)
	}
	return parsed.Session.Status
}
```
