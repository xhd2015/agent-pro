# Scenario

**Bug**: POST session follow-up messages must appear as user timeline entries

```
seed idle session -> POST .../messages {text} -> user event -> GET detail
```

## Preconditions

- Session exists under `AGENT_RUN_HOME` before follow-up POST.
- Web server uses Bearer `test`.

## Steps

1. Start web server.
2. Seed `fake-codex/follow-up-sess` session meta as `idle`.
3. POST `/api/agent-run/sessions/fake-codex/follow-up-sess/messages` with text `second prompt`.
4. `Run` GETs session detail.

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "web"
	req.WebTokenMode = "explicit"
	req.WebToken = "test"
	req.WebPort = 0
	req.SessionRunner = "fake-codex"
	req.SessionID = "follow-up-sess"
	req.CreatePrompt = "second prompt"
	startWebServer(t, req)

	if err := seedIdleSession(t, req.Home, req.SessionRunner, req.SessionID); err != nil {
		return err
	}

	url := req.WebBaseURL + "/api/agent-run/sessions/" + req.SessionRunner + "/" + req.SessionID + "/messages"
	payload, _ := json.Marshal(map[string]string{"text": req.CreatePrompt})
	status, body := httpPostJSON(t, url, req.WebToken, string(payload))
	if status != 200 && status != 202 {
		t.Fatalf("POST messages: status=%d body=%q", status, body)
	}

	req.HTTPMethod = "GET"
	req.HTTPAuth = req.WebToken
	req.HTTPPath = "/api/agent-run/sessions/" + req.SessionRunner + "/" + req.SessionID
	return nil
}

func seedIdleSession(t *testing.T, home, runner, sessionID string) error {
	t.Helper()
	sessDir := filepath.Join(home, "sessions", runner, sessionID)
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	meta := agentstorage.SessionMeta{
		Runner:    runner,
		SessionID: sessionID,
		Status:    "idle",
		CreatedAt: now,
		UpdatedAt: now,
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(sessDir, "meta.json"), metaBytes, 0644)
}
```