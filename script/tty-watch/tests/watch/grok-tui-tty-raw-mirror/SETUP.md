# Scenario

**Feature**: watch on a real TTY mirrors grok alternate-screen UI via raw terminal protocol

```
# grok-style alt-screen box UI with true-color SGR
harness PTY -> detached fake grok TUI -> watch PTY -> raw ESC sequences (not plain-text snapshot)
```

Regression from `renderObserverFrame`: watch converts PTY output to plain text even on
interactive terminals, so the observer cannot render the boxed grok layout like the
primary session. The observer TTY must receive raw escape sequences (`\x1b[?1049h`, etc.)
so its terminal emulator paints the same UI.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "watch-grok-tui-tty-raw-mirror"
	req.WatchProbe = "2s"
	return nil
}
```