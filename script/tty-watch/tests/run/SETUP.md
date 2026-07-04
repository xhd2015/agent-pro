# Scenario

**Feature**: `tty-watch run <cmd>` embeds ptywrap, attaches as writer, manages registry lifecycle

```
# writer attach path
harness PTY -> tty-watch run <cmd> -> embedded ptywrap -> registry session-N.json
# Ctrl-C -> child; Ctrl-] -> detach; exit -> prune registry
```

## Preconditions

- Isolated `TTY_WATCH_HOME` from root Setup.
- Default run command is `sleep` unless leaf overrides `RunCommand` (harness prepends `run`).

## Steps

1. Leaf sets `req.Phase` for the run scenario under test.
2. Harness starts tty-watch in PTY or detaches for registry assertions.
3. Assert checks registry state, host silence, or child-visible output.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("run setup: tty-watch binary not built (root Setup skipped?)")
	}
	return nil
}
```