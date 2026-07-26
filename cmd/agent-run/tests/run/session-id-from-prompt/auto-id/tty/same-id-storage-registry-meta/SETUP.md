# Scenario

**Feature**: TTY auto-id is identical across stderr, storage, meta, and registry

```
agent-run run --agent-runner grok-tty --keep-tty --session-id-from-prompt "hello world"
  -> stderr: grok-tty: hello-world-YYYYMMDD-HHMMSS
  -> sessions/grok-tty/<same-id>/
  -> meta.session_id == meta.terminal_session_id == <same-id>
  -> grok-tty-registry/<same-id>.json
```

## Steps

1. Run with prompt `hello world` (slug base `hello-world`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Prompt = "hello world"
	req.Args = append(req.Args, req.Prompt)
	return nil
}
```
