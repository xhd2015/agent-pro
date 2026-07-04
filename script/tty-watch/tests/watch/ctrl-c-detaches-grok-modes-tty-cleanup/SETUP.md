# Scenario

**Feature**: Ctrl-C detach restores observer terminal after grok mode preamble

```
# fake grok modes -> watch PTY -> iTerm kitty Ctrl-C (99;5u) -> terminal cleanup on stdout
harness -> detached sh grok-modes -> watch -> ESC[99;5u -> watch exits + restores TTY
```

Reproduces the post-detach corruption users see in iTerm2: watch exits after
`\x1b[99;5u`, but grok had already enabled alt-screen (`1049h`), kitty keyboard
push (`\x1b[?u`), and mouse tracking on the observer terminal. Cleanup must pop
with `\x1b[<u`, not only `\x1b[?0u`. Without that pop, the host shell shows
random escape garbage on later typing and mouse moves.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "watch-ctrl-c-detaches-grok-modes-tty-cleanup"
	return nil
}
```