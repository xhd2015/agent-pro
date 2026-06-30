# Scenario

**Bug**: web follow-up runs must pass conversation history so the agent recalls earlier user text

```
POST create "hi" -> wait finished -> POST messages "what did I ask?" -> assistant mentions "hi"
```

## Preconditions

- `fake-codex` on PATH; continuation prompt includes prior user `hi`.
- Web server uses explicit Bearer token.

## Steps

1. Leaf starts web server and performs two-turn HTTP flow before `Run` GET detail.
2. `Assert` polls until an assistant `message` references the first prompt.

```go
import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	req.Mode = "web"
	return nil
}

func postSessionMessage(t *testing.T, req *Request, runner, sessionID, text string) {
	t.Helper()
	url := req.WebBaseURL + "/api/agent-run/sessions/" + runner + "/" + sessionID + "/messages"
	payload, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		t.Fatalf("marshal message: %v", err)
	}
	status, body := httpPostJSON(t, url, req.WebToken, string(payload))
	if status != http.StatusAccepted && status != 200 && status != 202 {
		t.Fatalf("POST messages: status=%d body=%q", status, body)
	}
}

func assistantMessagesMentioning(detailJSON, substr string) bool {
	substr = strings.ToLower(strings.TrimSpace(substr))
	if substr == "" {
		return false
	}
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
		if ev["type"] != "message" || ev["role"] != "assistant" {
			continue
		}
		text, _ := ev["text"].(string)
		if strings.Contains(strings.ToLower(text), substr) {
			return true
		}
	}
	return false
}

func waitForAssistantMention(t *testing.T, req *Request, runner, sessionID, substr string, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		_, body := getSessionDetail(t, req, runner, sessionID)
		if assistantMessagesMentioning(body, substr) {
			return body
		}
		time.Sleep(100 * time.Millisecond)
	}
	_, body := getSessionDetail(t, req, runner, sessionID)
	t.Fatalf("timeout waiting for assistant mention of %q in: %s", substr, body)
	return body
}
```