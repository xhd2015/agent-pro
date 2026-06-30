# Scenario

**Feature**: mobile active chat — seeded session with messages

```
AGENT_RUN_HOME session seed → agent-run web → token + /session → chat-active + composer
```

## Preconditions

- `playwright-debug` on PATH (else `t.Skip`).
- Web server started with `--token test-token`.
- Session `fake-opencode/layout-chat` seeded with `meta.json` and `events.jsonl` before browser opens.

## Steps

1. Create session files under `req.Home/sessions/fake-opencode/layout-chat/`.
2. Start `agent-run web` in background on a free port.
3. Build playwright script: viewport 390×844, inject token, open session route, assert chat-active and composer.

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	req.Layout = "chat-active"
	req.WebTokenMode = "explicit"
	req.Token = "test-token"
	runner := "fake-opencode"
	sessionID := "layout-chat"

	if err := seedActiveSession(t, req.Home, runner, sessionID); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	sessionPath := "/sessions/" + runner + "/" + sessionID
	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + sessionPath + `', { waitUntil: 'networkidle' });
const chat = page.locator('[data-testid="chat-active"]');
await chat.waitFor({ state: 'visible', timeout: 15000 });
const messages = page.locator('[data-testid="message-list"]');
await messages.waitFor({ state: 'visible', timeout: 15000 });
const count = await messages.locator('[data-testid="message-item"]').count();
if (count < 1) throw new Error('expected at least one message-item, got ' + count);
` + assertComposerPinnedBottom()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}

func seedActiveSession(t *testing.T, home, runner, sessionID string) error {
	t.Helper()
	sessDir := filepath.Join(home, "sessions", runner, sessionID)
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	meta := agentstorage.SessionMeta{
		Runner:          runner,
		SessionID:       sessionID,
		RunnerSessionID: "inner-layout-chat",
		Status:          "idle",
		Model:           "test-model",
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(sessDir, "meta.json"), metaBytes, 0644); err != nil {
		return err
	}
	ev := types.AgentEvent{Type: types.ActionMessage, Text: "Hello from layout test"}
	evBytes, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	eventsPath := filepath.Join(sessDir, "events.jsonl")
	return os.WriteFile(eventsPath, append(evBytes, '\n'), 0644)
}
```