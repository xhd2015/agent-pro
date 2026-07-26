# Scenario

**Feature**: browser UI exposes tty runner choice and terminal modal without hiding chat content

```
home runner picker -> codex-tty + grok-tty options (fixture listing only)
seeded tty chat + fake ptywrap registry -> terminal icon -> modal websocket attach
modal close -> transcript remains; navigation back -> same terminal attach
```

## Preconditions

- `playwright-debug` is installed for UI leaves.
- UI leaves in this subtree use **seeded fixtures + fake ptywrap** only — no live
  `POST grok-tty` / `fake-codex` agent runs. Live grok mock harness lives in
  `web/tty-terminal-persistent/frontend-ui/` and `web-cli-subset/frontend-ui/`.

## Steps

1. Leaf setup seeds any required session and tty registry fixture.
2. Leaf setup writes Playwright script.
3. `Run` executes the script.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "ui"
	return nil
}
```
