# Scenario

**Bug**: on mobile home, long workspace path pushes runner picker off-screen

```
web cwd = deep nested dir -> GET status.workspace is long -> / (open API) -> runner-picker within 390px viewport
```

## Preconditions

- `playwright-debug` on PATH.
- `agent-run web` process working directory is a deeply nested path under `t.TempDir()`.
- Open API mode (no `--token`); no localStorage token.

## Steps

1. Create deep workspace directory and set `req.WebWorkingDir`.
2. Start `agent-run web` in background without `--token`.
3. Open `/`, wait for workspace + runner controls; assert picker fits viewport width.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	req.Layout = "home-long-workspace"
	req.WebTokenMode = "omit"
	req.WebWorkingDir = makeDeepWorkspaceDir(t, req.TempDir)

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := `
await page.addInitScript(() => {
  localStorage.removeItem('agent-run-token');
});
await page.goto('` + req.BaseURL + `/', { waitUntil: 'networkidle' });
const workspace = page.locator('[data-testid="workspace"]');
await workspace.waitFor({ state: 'visible', timeout: 15000 });
const wsText = await workspace.innerText();
if (!wsText || wsText.length < 40) {
  throw new Error('expected long workspace path text, got: ' + wsText);
}
const empty = page.locator('[data-testid="empty-state"]');
await empty.waitFor({ state: 'visible', timeout: 15000 });
` + assertRunnerPickerWithinViewport() + assertComposerPinnedBottom()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```