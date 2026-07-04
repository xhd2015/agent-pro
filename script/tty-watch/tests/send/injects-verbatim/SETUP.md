# Scenario

**Feature**: send injects multi-word message bytes exactly as provided

```
# detached cat capture; send follow-up text
harness -> detached cat capture -> tty-watch send session-N "follow up" -> capture.bin == follow up
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "send-injects-verbatim"
	req.SendMessage = "follow up"
	return nil
}
```