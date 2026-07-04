# Scenario

**Feature**: watch replicates grok-like interactive TUI without control-character garbage

```
# grok-style alternate-screen prompt with cursor hide/show toggles
harness -> detached fake grok TUI -> watch -> clean prompt text, no CSI/C0 leaks
```

Reproduces stray control characters in the input area when observing a `tty-watch run grok`
session: grok enters the alternate screen (`\x1b[?1049h`) and toggles cursor visibility
(`\x1b[?25l` / `\x1b[?25h`) around the prompt. Watch must mirror the visible session,
not dump raw PTY escape sequences into the observer terminal.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "watch-grok-like-prompt"
	req.WatchProbe = "2s"
	return nil
}
```