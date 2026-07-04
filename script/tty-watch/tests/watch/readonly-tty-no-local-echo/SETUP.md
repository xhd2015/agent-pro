# Scenario

**Feature**: watch on a TTY silently drops local keyboard and mouse input

```
# detached echo loop; watch on PTY; harness types + sends mouse sequence
harness PTY -> watch -> typed probe and mouse CSI must not appear on screen
```

Bug: `streamObserver` never puts stdin in raw mode or drains input on a real TTY,
so the terminal driver locally echoes typed characters and mouse-report escape
sequences onto the watching screen. Watch is readonly — all input except Ctrl-C
must be silently dropped with no local echo.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "watch-readonly-tty-no-local-echo"
	return nil
}
```