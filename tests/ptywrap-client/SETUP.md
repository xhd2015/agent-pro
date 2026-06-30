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

- `requires-tty` leaf uses pipe stdin (non-TTY) via root helper or env override.

```go
import (
	"os"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	r, w, err := os.Pipe()
	if err != nil {
		return err
	}
	t.Cleanup(func() {
		r.Close()
		w.Close()
	})
	// Save original; attach leaf with Phase attach-requires-tty swaps to pipe.
	t.Setenv("PTYWRAP_CLIENT_TEST_STDIN_FD", "pipe")
	req.UsePipeStdin = true
	return nil
}
```