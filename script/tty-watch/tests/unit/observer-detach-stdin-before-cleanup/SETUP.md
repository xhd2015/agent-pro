# Scenario

**Feature**: Observer detach restores stdin termios before writing stdout cleanup

```
# read attach.go -> detachCleanup must Restore stdin before writeObserverTTYDetachCleanup
```

User report after `tty-cleanup` fix: iTerm2 shows `^[[?0u` on the prompt and subsequent
typing produces kitty garbage (`0u9;5:3u`, `e1;1:3uc9;1:3u`). Cleanup CSI was written
while stdin was still in raw mode (`defer term.Restore`), so iTerm echoes/displays the
disable sequence and kitty keyboard mode stays active.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "unit-observer-detach-stdin-before-cleanup"
	return nil
}
```