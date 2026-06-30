# Scenario

**Bug**: follow-up via composer must not duplicate user timeline bubbles during agent run

```
seed idle session + 1 user event -> session page -> composer follow-up -> fake-codex run -> exactly 2 user bubbles
```

## Preconditions

- `fake-codex` on PATH (built into temp `bin/`).
- `playwright-debug` on PATH.
- Seeded session `status=idle` with one prior user message in `events.jsonl`.

## Steps

1. Seed `fake-codex/follow-up-dedupe` with initial user prompt `first layout prompt`.
2. Start `agent-run web` with open API on a free port.
3. Open session route; send `second follow-up prompt` via composer.
4. While session is running, poll user bubble count every 250ms; fail immediately if count > 2.
5. After run completes, assert exactly two user bubbles; each prompt text appears once.

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

	req.Layout = "follow-up-dedupe"
	req.WebTokenMode = "omit"
	runner := "fake-codex"
	sessionID := "follow-up-dedupe"
	firstPrompt := "first layout prompt"
	followUpPrompt := "second follow-up prompt"

	if err := seedIdleSessionWithUserMessage(t, req.Home, runner, sessionID, firstPrompt); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	sessionPath := "/sessions/" + runner + "/" + sessionID
	body := `
await page.goto('` + req.BaseURL + sessionPath + `', { waitUntil: 'domcontentloaded' });
const chat = page.locator('[data-testid="chat-active"]');
await chat.waitFor({ state: 'visible', timeout: 15000 });
` + assertUserMessageCount(1) + sendComposerMessage(followUpPrompt) + assertNoDuplicateUserMessagesDuringRun(2) +
		waitForSessionRunComplete() + assertUserMessageCount(2) + assertDistinctUserPromptsOnce([]string{firstPrompt, followUpPrompt})

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```