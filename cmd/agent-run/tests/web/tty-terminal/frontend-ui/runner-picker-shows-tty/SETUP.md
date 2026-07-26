# Scenario

**Feature**: home runner picker includes tty runners

```
GET runners -> home runner select -> options include codex-tty and grok-tty
```

## Preconditions

- Home page loads runners from typed frontend API helpers.

## Steps

1. Open home page with valid token.
2. Inspect runner select options.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.PlaywrightScript = browserScript(req, `
const select = page.locator('[data-testid="runner-select"]');
await select.waitFor({ state: 'visible', timeout: 15000 });
const options = await select.locator('option').allTextContents();
if (!options.some((text) => text.includes('codex-tty'))) throw new Error('missing codex-tty option: ' + options.join(','));
if (!options.some((text) => text.includes('grok-tty'))) throw new Error('missing grok-tty option: ' + options.join(','));
`)
	return nil
}
```
