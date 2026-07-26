# Scenario

**Feature**: ptywrap client library talks to daemon over HTTP and WS

```
# client stack
CLI/test -> ptywrap/client -> HTTP list/create + WS attach bridge
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap/client` is importable.
- Implementer adds test hooks as needed:
  - `AttachWithIO` accepting explicit `io.Reader`/`io.Writer` fds (or equivalent).
  - `Client.SetTestSessions` or injectable list func for resolve tests without daemon.
  - Fake TTY helper (`startPTYPair`) for attach-captures-id leaf.

## Context

- `requires-tty` leaf sets `req.UsePipeStdin` / `UsePipeStdout` explicitly (no process env).

```go
import (
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// No os.Setenv: pipe non-TTY is selected via req.UsePipeStdin on the leaf that
	// needs it (attach/requires-tty). Product harness must honor Request fields.
	return nil
}
```
