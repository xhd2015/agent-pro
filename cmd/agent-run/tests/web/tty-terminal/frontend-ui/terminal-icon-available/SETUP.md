# Scenario

**Feature**: tty chat page shows terminal icon when terminal is available

```
codex-tty chat + live registry -> top bar terminal button with accessible label
```

## Preconditions

- Terminal status endpoint returns `available: true`.

## Steps

1. Seed tty session and live registry.
2. Open session page.
3. Wait for terminal button.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Runner = "codex-tty"
	req.SessionID = "ui-terminal-available"
	req.RegistryTranscript = "ui-terminal-ready\n"
	listenAddr := startFakePtywrap(t, req)
	writeSessionFixture(t, req, req.Runner, req.SessionID, "running")
	writeTTYRegistryFixture(t, req, req.Runner, req.SessionID, listenAddr)
	req.PlaywrightScript = sessionBrowserScript(req, `
const terminalButton = page.getByRole('button', { name: /terminal/i });
await terminalButton.waitFor({ state: 'visible', timeout: 15000 });
if (!(await terminalButton.isEnabled())) throw new Error('terminal button is disabled despite live registry');
`)
	return nil
}
```
