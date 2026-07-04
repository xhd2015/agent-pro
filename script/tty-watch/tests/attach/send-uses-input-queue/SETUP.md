# Scenario

**Feature**: `tty-watch send` shares the serialized input queue with attach writers

```
# attach connected; send injects marker; PTY child capture matches without corruption
harness -> detached cat capture -> attach holds session -> send marker -> capture.bin exact
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "attach-send-input-queue"
	req.SendMessage = "ATTACH_SEND_QUEUE_MARKER"
	return nil
}
```