# Scenario

**Feature**: thinking progress card body renders markdown (not plain pre-wrap only)

```
seed flat session with think text containing **bold** and `inline code`
  → /sessions/<id> → thinking progress-card body has strong and/or code
```

## Preconditions

- `playwright-debug` on PATH.
- Flat seed with `type=think` markdown markers.
- Open API.
- Label: `ui-automation`.

## Steps

1. Seed think + user + short assistant events.
2. Open session; locate thinking `progress-card`.
3. Assert body has `strong` or `code` for markdown markers.

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "thinking-markdown-renders"
	req.WebTokenMode = "omit"
	runner := "fake-opencode"
	sessionID := "layout-md-thinking"
	thinkText := "Plan: run **pwd** then `ls` and format results for the user."

	if err := seedMarkdownThinkingSession(t, req.Home, runner, sessionID, thinkText); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	sessionPath := "/sessions/" + sessionID
	body := `
await page.goto('` + req.BaseURL + sessionPath + `', { waitUntil: 'networkidle', timeout: 45000 });
await page.waitForSelector('[data-testid="message-list"], [data-testid="chat-active"]', { timeout: 20000 });
const thinkCard = page.locator('[data-testid="progress-card"]').filter({
  has: page.locator('.progress-card-label', { hasText: /thinking/i }),
}).first();
await thinkCard.waitFor({ state: 'visible', timeout: 20000 });
const snap = await thinkCard.evaluate((card) => {
  const body = card.querySelector('.progress-card-body') || card;
  const text = (body.textContent || '').trim();
  return {
    text: text.slice(0, 300),
    hasStrong: !!body.querySelector('strong, b'),
    hasCode: !!body.querySelector('code'),
    hasPre: !!body.querySelector('pre'),
    hasMarkdownRoot: !!body.querySelector('[data-testid="markdown-body"], .markdown-body'),
  };
});
if (!snap.hasStrong && !snap.hasCode) {
  throw new Error(
    'thinking card has markdown markers but no <strong>/<code> structure: ' +
      JSON.stringify(snap),
  );
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}

func seedMarkdownThinkingSession(t *testing.T, home, runner, sessionID, thinkText string) error {
	t.Helper()
	sessDir := filepath.Join(home, "sessions", sessionID)
	if err := os.MkdirAll(sessDir, 0755); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	meta := map[string]any{
		"runner":     runner,
		"session_id": sessionID,
		"status":     "idle",
		"workspace":  "/tmp/demo-workspace",
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
	ts := time.Now().UTC().Add(-3 * time.Minute).UnixMilli()
	events := []map[string]any{
		{"type": "message", "role": "user", "timestamp": ts, "text": "run ls and pwd"},
		{"type": "think", "timestamp": ts + 500, "text": thinkText},
		{"type": "message", "role": "assistant", "timestamp": ts + 1000, "text": "Done."},
	}
	var lines []byte
	for _, ev := range events {
		raw, err := json.Marshal(ev)
		if err != nil {
			return err
		}
		lines = append(lines, raw...)
		lines = append(lines, '\n')
	}
	return os.WriteFile(filepath.Join(sessDir, "events.jsonl"), lines, 0644)
}
```
