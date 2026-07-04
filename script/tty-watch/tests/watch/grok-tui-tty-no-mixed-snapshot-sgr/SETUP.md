# Scenario

**Feature**: watch TTY must not mix plain-text screen snapshot with raw SGR incremental updates

```
# ptywrap scrollback replay (?25l snapshot) + grok-style true-color input animation
harness PTY -> detached fake grok TUI -> watch PTY -> no [38;2; garbage in input area
```

Regression: on TTY observer (`observerMode=false`), the first `isScreenSnapshotFrame`
is still converted to plain text while later incremental `\x1b[38;2;...` frames pass
through raw. The observer terminal never enters alternate-screen mode, so true-color
SGR sequences render as visible `[38;2;...` garbage in the input box (user screenshot).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "watch-grok-tui-tty-no-mixed-snapshot-sgr"
	req.WatchProbe = "3s"
	return nil
}
```