# Scenario

**Feature**: `tty-watch attach` joins a live session as multi-writer attacher (write+resize)

```
# attach_mode=attach: full write+resize, multiplexed output, serialized input queue
harness -> detached session -> tty-watch attach <id> -> WS attach_mode=attach -> PTY master
PTY read loop -> broadcast -> attachers + observers + screen writer
stdin/resize/send -> inputCh -> single input loop per session
```

## Preconditions

- Leaves start a detached session (or attached `run` for co-writer cases) before invoking `attach`.
- Multi-client write leaves may use `ttyrunner.DialPTYAttach(..., "attach")` when two concurrent
  CLI attach processes are awkward; primary connect/detach/error leaves use `tty-watch attach`.
- Detach uses Ctrl-] (`\x1d`), same as `run`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Bin == "" {
		t.Fatalf("attach setup: tty-watch binary not built")
	}
	return nil
}
```