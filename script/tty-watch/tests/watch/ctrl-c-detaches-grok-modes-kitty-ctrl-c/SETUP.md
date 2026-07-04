# Scenario

**Feature**: Ctrl-C detaches watch after grok terminal mode preamble (kitty encoding)

```
# fake grok mode preamble -> watch PTY -> kitty Ctrl-C -> watch detaches
harness -> detached sh grok-modes -> watch -> CSI 3;5 u -> detach exit 0
```

Synthetic reproduction of the real-grok bug without requiring grok on PATH: mirrors
the terminal mode sequences grok emits (`1049h`, mouse tracking, `2004h`, `?u`),
then sends the kitty keyboard protocol Ctrl-C bytes a real terminal produces.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "watch-ctrl-c-detaches-grok-modes-kitty-ctrl-c"
	return nil
}
```