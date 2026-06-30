# Scenario

**Bug**: live session updates must use SSE tail, not aggressive session-detail polling

```
init transport monitor -> create fake-codex session -> session page streams 8s -> SSE >=1, detail GET <=3
```

## Preconditions

- `fake-codex` on PATH (built into temp `bin/`).
- `playwright-debug` on PATH.
- Open API (`WebTokenMode=omit`).

## Steps

1. Build `fake-codex`; start `agent-run web` on free port.
2. Register `page.on('request')` counters before navigation.
3. Create live session via API; open session page; monitor network for 8s while assistant streams.
4. Assert ≥1 `.../events/stream` request and ≤3 session-detail GETs (exclude stream path).

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

	req.Layout = "sse-transport"
	req.WebTokenMode = "omit"
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	createJS := initStreamingTransportMonitor() + `
const res = await fetch('` + req.BaseURL + `/api/agent-run/sessions', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ runner: 'fake-codex', prompt: 'monitor sse transport' }),
});
if (!res.ok && res.status !== 202) throw new Error('create session failed: ' + res.status);
const data = await res.json();
const sid = data.session.session_id;
const runner = data.session.runner;
await page.goto('` + req.BaseURL + `/sessions/' + runner + '/' + sid, { waitUntil: 'domcontentloaded' });
const chat = page.locator('[data-testid="chat-active"]');
await chat.waitFor({ state: 'visible', timeout: 15000 });
` + assertStreamingTransportProfile()

	req.PlaywrightScript = mobileViewportScript(createJS)
	return nil
}
```