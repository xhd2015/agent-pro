# Scenario

**Feature**: session detail UI shows workspace path and user message bubbles

```
seed session(workspace, user+assistant events) -> web -> playwright session route -> workspace + message-item-user
```

## Preconditions

- `playwright-debug` on PATH.
- Seeded `events.jsonl` includes `role=user` and `role=assistant` message events.
- `meta.json` includes `workspace` path displayed in session header.

## Steps

1. Seed `fake-opencode/layout-workspace` with workspace and role-tagged events.
2. Start `agent-run web` with explicit token.
3. Run playwright script opening session URL.

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	req.Layout = "workspace-session"
	req.WebTokenMode = "explicit"
	req.Token = "test-token"
	runner := "fake-opencode"
	sessionID := "layout-workspace"
	workspacePath := filepath.Join(req.TempDir, "demo-workspace")

	if err := seedWorkspaceSession(t, req.Home, runner, sessionID, workspacePath); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	escapedWS := strings.ReplaceAll(workspacePath, `\`, `\\`)
	escapedWS = strings.ReplaceAll(escapedWS, `'`, `\'`)
	sessionPath := "/sessions/" + runner + "/" + sessionID
	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + sessionPath + `', { waitUntil: 'networkidle' });
const workspace = page.locator('[data-testid="workspace"]');
await workspace.waitFor({ state: 'visible', timeout: 15000 });
const wsText = await workspace.innerText();
if (!wsText.includes('` + escapedWS + `')) {
  throw new Error('workspace text missing path: ' + wsText);
}
const userMsg = page.locator('[data-testid="message-item-user"]');
await userMsg.waitFor({ state: 'visible', timeout: 15000 });
const userText = await userMsg.innerText();
if (!userText.includes('fix the layout bug')) {
  throw new Error('user message text mismatch: ' + userText);
}
const assistantMsg = page.locator('[data-testid="message-item-assistant"]');
await assistantMsg.waitFor({ state: 'visible', timeout: 15000 });
` + assertComposerPinnedBottom()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}

func seedWorkspaceSession(t *testing.T, home, runner, sessionID, workspace string) error {
	t.Helper()
	sessDir := filepath.Join(home, "sessions", runner, sessionID)
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	meta := map[string]any{
		"runner":     runner,
		"session_id": sessionID,
		"status":     "idle",
		"workspace":  workspace,
		"created_at": now,
		"updated_at": now,
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(sessDir, "meta.json"), metaBytes, 0644); err != nil {
		return err
	}
	eventsNDJSON := "{\"type\":\"message\",\"role\":\"user\",\"text\":\"fix the layout bug\"}\n" +
		"{\"type\":\"message\",\"role\":\"assistant\",\"text\":\"On it.\"}\n"
	return os.WriteFile(filepath.Join(sessDir, "events.jsonl"), []byte(eventsNDJSON), 0644)
}
```