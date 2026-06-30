# Scenario

**Feature**: mobile auth page — missing token shows token entry

```
agent-run web (background) → / without token → auth-page, no horizontal scroll
```

## Preconditions

- `playwright-debug` on PATH (else `t.Skip`).
- Web server started with `--token test-token`.
- Browser storage does **not** contain `agent-run-token`.

## Steps

1. Allocate free port and start `agent-run web` in background.
2. Build playwright script: viewport 390×844, clear storage, open `/`, assert auth-page layout.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	req.Layout = "auth"
	req.WebTokenMode = "explicit"
	req.Token = "test-token"
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
await auth.waitFor({ state: 'visible', timeout: 15000 });
` + assertAuthPagePinnedBottom()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```