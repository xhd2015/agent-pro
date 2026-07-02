# Scenario

**Feature**: browser UI exposes tty runner choice and terminal modal without hiding chat content

```
home runner picker -> codex-tty + grok-tty options
tty chat + available terminal -> terminal icon -> modal websocket attach
modal close -> transcript remains; navigation back -> same terminal attach
```

## Preconditions

- `playwright-debug` is installed for UI leaves.
- UI API calls stay in frontend API helpers; components consume typed helpers.

## Steps

1. Leaf setup seeds any required session and tty registry fixture.
2. Leaf setup writes Playwright script.
3. `Run` executes the script.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "ui"
	return nil
}
```
