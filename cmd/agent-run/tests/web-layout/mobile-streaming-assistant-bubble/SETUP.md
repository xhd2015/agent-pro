# Scenario

**Feature**: assistant reply streams into a single bubble with growing visible text

```
POST create on session page -> phased assistant events -> message-item-assistant text length increases
```

## Preconditions

- `fake-codex` on PATH (root layout build includes agent-run only — run uses API create from browser).
- Open API or token in localStorage; live run emits streaming phases.

## Steps

1. Start web with open API (`WebTokenMode=omit`) on free port.
2. Navigate to home, create session via UI or API from playwright (use fetch from page).
3. Poll assistant bubble text until length increases.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	fakeCodex := filepath.Join(req.TempDir, "bin", "fake-codex")
	if err := os.MkdirAll(filepath.Dir(fakeCodex), 0755); err != nil {
		return err
	}
	build := exec.Command("go", "build", "-o", fakeCodex, "./cmd/fake-codex")
	build.Dir = req.RepoRoot
	if out, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("build fake-codex: %w\n%s", err, string(out))
	}
	req.Env = append(req.Env, "PATH="+filepath.Dir(fakeCodex)+string(os.PathListSeparator)+os.Getenv("PATH"))

	req.Layout = "streaming-bubble"
	req.WebTokenMode = "omit"
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	createJS := `
const res = await fetch('` + req.BaseURL + `/api/agent-run/sessions', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ runner: 'fake-codex', prompt: 'grow the stream' }),
});
if (!res.ok && res.status !== 202) throw new Error('create session failed: ' + res.status);
const data = await res.json();
const sid = data.session.session_id;
const runner = data.session.runner;
await page.goto('` + req.BaseURL + `/sessions/' + runner + '/' + sid, { waitUntil: 'domcontentloaded' });
`
	body := createJS + assertAssistantBubbleTextGrows()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```