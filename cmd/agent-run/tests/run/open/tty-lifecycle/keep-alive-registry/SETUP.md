# Scenario

**Feature**: `--open` implies keep-alive — registry entry remains reachable after open completes

```
agent-run run --agent-runner grok-tty --open "hi"
  -> exit 0 after attach returns
  -> grok-tty-registry/<id>.json exists
  -> listen_addr TCP still open (reattach/send possible)
```

## Preconditions

- Instant attach so the CLI process can finish while leaving the PTY server up
  (same product intent as keep-tty after detach).

## Steps

1. Run open lifecycle to completion.
2. Mode `open-registry-after` loads registry for the printed session id.
3. Assert registry file + TCP reachability.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Prompt = "hi"
	req.Mode = "open-registry-after"
	req.Args = []string{"run", "--agent-runner", "grok-tty", "--open", req.Prompt}
	// Hold long enough that keep-alive server outlives the CLI open path if the
	// child TUI would otherwise exit immediately after prompt inject.
	setGrokTTYCommand(req, fakeTUIHoldSeconds(30))
	t.Cleanup(func() {
		// Best-effort: remove registry leftovers under temp home (temp dir cleans).
	})
	return nil
}
```
