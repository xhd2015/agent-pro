# Scenario

**Feature**: mobile empty state — open API, no auth gate, no sessions

```
agent-run web without --token (background) → / → empty-state (not auth-page) + composer pinned bottom
```

## Preconditions

- `playwright-debug` on PATH (else `t.Skip`).
- Web server started **without** `--token` (`WebTokenMode = omit`).
- No `agent-run-token` in localStorage; no sessions seeded under home.

## Steps

1. Allocate free port and start `agent-run web` in background (open API).
2. Build playwright script: viewport 390×844, clear storage, open `/`, assert empty-state and composer.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	req.Layout = "empty"
	req.WebTokenMode = "omit"
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := `
await page.addInitScript(() => {
  localStorage.removeItem('agent-run-token');
});
await page.goto('` + req.BaseURL + `/', { waitUntil: 'networkidle' });
const auth = page.locator('[data-testid="auth-page"]');
if (await auth.isVisible()) throw new Error('auth page must not show when API is open');
const empty = page.locator('[data-testid="empty-state"]');
await empty.waitFor({ state: 'visible', timeout: 15000 });
` + assertComposerPinnedBottom()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```