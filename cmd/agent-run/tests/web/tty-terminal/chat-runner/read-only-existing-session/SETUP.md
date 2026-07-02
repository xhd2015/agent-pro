# Scenario

**Feature**: existing chat page runner cannot be changed

```
/sessions/codex-tty/<id> -> runner metadata visible
no enabled runner select on session page -> follow-up bound to route runner
```

## Preconditions

- Parent setup seeded an existing `codex-tty` session.

## Steps

1. Open existing session page.
2. Verify runner metadata is visible and runner select is absent or disabled.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.PlaywrightScript = sessionBrowserScript(req, `
await page.getByText('codex-tty').waitFor({ state: 'visible', timeout: 15000 });
const enabledSelects = await page.locator('[data-testid="runner-select"]:enabled').count();
if (enabledSelects > 0) {
  throw new Error('existing chat page exposes enabled runner select');
}
const transcript = page.getByText('assistant keeps transcript');
await transcript.waitFor({ state: 'visible', timeout: 15000 });
`)
	return nil
}
```
