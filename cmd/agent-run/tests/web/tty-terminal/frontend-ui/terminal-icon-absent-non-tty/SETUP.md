# Scenario

**Feature**: non-tty chat page does not imply terminal attach availability

```
codex chat -> no enabled terminal attach button
```

## Preconditions

- Non-tty session metadata exists.

## Steps

1. Seed `codex` session.
2. Open session page.
3. Verify terminal button is absent or disabled.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "codex"
	req.SessionID = "ui-non-tty"
	writeSessionFixture(t, req, req.Runner, req.SessionID, "finished")
	req.PlaywrightScript = sessionBrowserScript(req, `
await page.getByText('assistant keeps transcript').waitFor({ state: 'visible', timeout: 15000 });
const enabled = await page.getByRole('button', { name: /terminal/i }).filter({ hasNotText: /unavailable/i }).count();
if (enabled > 0) throw new Error('non-tty chat exposes terminal attach button');
`)
	return nil
}
```
