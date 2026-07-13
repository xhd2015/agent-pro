# Scenario

**Feature**: assistant message body renders markdown (strong / pre / code)

```
seed flat session with assistant text containing **bold** + fenced code
  → /sessions/<id> → [data-testid=assistant-message] has strong and pre/code
  → textContent does not keep raw ** as sole bold representation
```

## Preconditions

- `playwright-debug` on PATH.
- Flat seed `sessions/layout-md-assistant/{meta.json,events.jsonl}`.
- Open API.
- Label: `ui-automation`.

## Steps

1. Seed assistant event with bold + fenced code markdown.
2. Start web; open flat session route.
3. Assert DOM has `strong` and `pre`/`code`; no raw `**` in textContent for bold markers.

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	req.Layout = "assistant-markdown-renders"
	req.WebTokenMode = "omit"
	runner := "fake-opencode"
	sessionID := "layout-md-assistant"
	userText := "run ls and pwd"
	// Build fenced markdown without embedding triple-backtick in this SETUP.md fence.
	fence := string([]byte{0x60, 0x60, 0x60})
	assistantText := "**pwd:** `/tmp/demo-workspace`\n\n**ls:**\n" + fence + "\nfile_a.md\nfile_b.go\nsubdir\n" + fence + "\n\nDone."

	if err := seedMarkdownAssistantSession(t, req.Home, runner, sessionID, userText, assistantText); err != nil {
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
const asst = page.locator('[data-testid="assistant-message"]').first();
await asst.waitFor({ state: 'visible', timeout: 20000 });
const snap = await asst.evaluate((el) => {
  const text = (el.textContent || '').trim();
  return {
    text: text.slice(0, 400),
    hasStrong: !!el.querySelector('strong, b'),
    hasPre: !!el.querySelector('pre'),
    hasCode: !!el.querySelector('code'),
    hasMarkdownRoot: !!el.querySelector('[data-testid="markdown-body"], .markdown-body'),
    literalStars: text.includes('**'),
    literalFence: text.includes('\u0060\u0060\u0060'),
  };
});
if (!snap.hasStrong) {
  throw new Error('assistant missing <strong> for bold markdown: ' + JSON.stringify(snap));
}
if (!snap.hasPre && !snap.hasCode) {
  throw new Error('assistant missing <pre>/<code> for fenced/inline code: ' + JSON.stringify(snap));
}
if (snap.literalStars) {
  throw new Error('assistant still shows raw ** markers: ' + JSON.stringify(snap));
}
if (snap.literalFence) {
  throw new Error('assistant still shows raw fence markers: ' + JSON.stringify(snap));
}
`
	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}

func seedMarkdownAssistantSession(t *testing.T, home, runner, sessionID, userText, assistantText string) error {
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
		{"type": "message", "role": "user", "timestamp": ts, "text": userText},
		{"type": "message", "role": "assistant", "timestamp": ts + 1000, "text": assistantText},
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
