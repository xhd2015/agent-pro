# Scenario

**Feature**: Ctrl-C detaches watch on real grok alt-screen TUI (kitty keyboard encoding)

```
# real grok -> detach -> watch PTY -> grok enables ?u + 1049h -> kitty Ctrl-C -> watch detaches
harness -> tty-watch run grok -> detach -> watch -> CSI 3;5 u -> detach exit 0
```

Bug: grok enables kitty keyboard protocol (`\x1b[?u`) and alternate screen on the
observer TTY. After that, a real Ctrl-C is encoded as `\x1b[3;5u`, not byte `\x03`
and not SIGINT. `drainObserverInput` only recognizes `\x03`, so watch appears stuck
while `bash --login -i` (no kitty protocol) still detaches on Ctrl-C.

```go
import (
	"os/exec"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if _, err := exec.LookPath("grok"); err != nil {
		t.Skip("grok not found in PATH")
	}
	req.Phase = "watch-ctrl-c-detaches-real-grok-kitty-ctrl-c"
	return nil
}
```