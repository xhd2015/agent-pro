# Scenario

**Subcommand**: `web` — localhost HTTP API; auth depends on `--token` mode

```
# omit --token → open API + startup warning
agent-run web --port 0 → GET health (no Bearer) → 200

# explicit --token <value> → Bearer required
agent-run web --token test → GET health (no Bearer) → 401

# --token auto → generated token on stderr + Bearer required
agent-run web --token auto → stderr token line; Bearer required
```

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).
- Web server binds `127.0.0.1` only.
- Leaves set `req.WebTokenMode` (`omit` | `explicit` | `auto`) and `req.WebPort` before `startWebServer`.

## Steps

1. Leaf `Setup` sets `req.Mode = "web"`, token mode, port, and HTTP probe fields.
2. Leaf `Setup` calls `startWebServer` to launch `agent-run web` in the background.
3. `Run` performs `httpGet` against the health endpoint (with or without Bearer).
4. `Assert` checks HTTP status, stderr, or filesystem side effects.

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

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Mode != "" && req.Mode != "web" {
		return fmt.Errorf("web group: unexpected Mode %q", req.Mode)
	}
	if req.WebTokenMode == "" {
		req.WebTokenMode = "explicit"
	}
	if req.WebTokenMode == "explicit" && req.WebToken == "" {
		req.WebToken = "test"
	}
	return nil
}

func postCreateSession(t *testing.T, req *Request, runner, prompt string) (string, int, string) {
	t.Helper()
	url := req.WebBaseURL + "/api/agent-run/sessions"
	payload, err := json.Marshal(map[string]string{
		"runner": runner,
		"prompt": prompt,
	})
	if err != nil {
		t.Fatalf("marshal create session: %v", err)
	}
	status, body := httpPostJSON(t, url, req.WebToken, string(payload))
	if status != http.StatusAccepted && status != 200 {
		t.Fatalf("POST sessions: status=%d body=%q", status, body)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("parse create response: %v body=%q", err, body)
	}
	sess, _ := parsed["session"].(map[string]any)
	if sess == nil {
		t.Fatalf("missing session in response: %q", body)
	}
	id, _ := sess["session_id"].(string)
	if strings.TrimSpace(id) == "" {
		t.Fatalf("empty session_id in response: %q", body)
	}
	return id, status, body
}

func getSessionDetail(t *testing.T, req *Request, runner, sessionID string) (int, string) {
	t.Helper()
	url := fmt.Sprintf("%s/api/agent-run/sessions/%s/%s", req.WebBaseURL, runner, sessionID)
	return httpGet(t, url, req.WebToken)
}

func eventsContainUserText(detailJSON, wantText string) bool {
	var parsed map[string]any
	if err := json.Unmarshal([]byte(detailJSON), &parsed); err != nil {
		return false
	}
	events, _ := parsed["events"].([]any)
	for _, raw := range events {
		ev, _ := raw.(map[string]any)
		if ev == nil {
			continue
		}
		role, _ := ev["role"].(string)
		text, _ := ev["text"].(string)
		typ, _ := ev["type"].(string)
		if typ == "message" && role == "user" && text == wantText {
			return true
		}
	}
	return false
}

func sessionStatusFromDetail(detailJSON string) string {
	var parsed map[string]any
	if err := json.Unmarshal([]byte(detailJSON), &parsed); err != nil {
		return ""
	}
	sess, _ := parsed["session"].(map[string]any)
	if sess == nil {
		return ""
	}
	status, _ := sess["status"].(string)
	return status
}

func waitForSessionStatus(t *testing.T, req *Request, runner, sessionID, wantStatus string, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, body := getSessionDetail(t, req, runner, sessionID)
		if sessionStatusFromDetail(body) == wantStatus {
			return body
		}
		time.Sleep(50 * time.Millisecond)
	}
	_, body := getSessionDetail(t, req, runner, sessionID)
	t.Fatalf("timeout waiting for session status %q, got %q: %s", wantStatus, sessionStatusFromDetail(body), body)
	return body
}

func messageTimestampForRole(detailJSON, role string) (float64, bool) {
	var parsed map[string]any
	if err := json.Unmarshal([]byte(detailJSON), &parsed); err != nil {
		return 0, false
	}
	events, _ := parsed["events"].([]any)
	for _, raw := range events {
		ev, _ := raw.(map[string]any)
		if ev == nil {
			continue
		}
		typ, _ := ev["type"].(string)
		r, _ := ev["role"].(string)
		if typ != "message" || r != role {
			continue
		}
		ts, ok := ev["timestamp"].(float64)
		if !ok || ts <= 0 {
			continue
		}
		return ts, true
	}
	return 0, false
}

func waitForMessageTimestamp(t *testing.T, req *Request, runner, sessionID, role string, timeout time.Duration) (string, float64) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, body := getSessionDetail(t, req, runner, sessionID)
		if ts, ok := messageTimestampForRole(body, role); ok {
			return body, ts
		}
		time.Sleep(50 * time.Millisecond)
	}
	_, body := getSessionDetail(t, req, runner, sessionID)
	t.Fatalf("timeout waiting for message timestamp role=%q in: %s", role, body)
	return body, 0
}

func waitForUserEvent(t *testing.T, req *Request, runner, sessionID, wantText string, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, body := getSessionDetail(t, req, runner, sessionID)
		if eventsContainUserText(body, wantText) {
			return body
		}
		time.Sleep(50 * time.Millisecond)
	}
	_, body := getSessionDetail(t, req, runner, sessionID)
	t.Fatalf("timeout waiting for user event text=%q in detail: %s", wantText, body)
	return body
}
```