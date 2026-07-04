# Scenario

**Feature**: tty-watch unit-level conversion helpers mirror writer attach output shaping

```
scrollback bytes -> ptywrap screen snapshot frame -> tty-watch screenSnapshotToText
```

## Preconditions

- Root Setup builds `tty-watch` and sets isolated `TTY_WATCH_HOME` (unused by unit leaves).

## Steps

1. Leaf sets `req.Phase` for the unit scenario under test.
2. Harness exercises scrollback/snapshot conversion without a live PTY session.
3. Assert checks column-zero layout and exit-marker alignment.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("unit setup: tty-watch binary not built (root Setup skipped?)")
	}
	return nil
}
```