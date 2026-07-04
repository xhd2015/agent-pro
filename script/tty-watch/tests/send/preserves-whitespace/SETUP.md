# Scenario

**Feature**: send preserves leading and trailing whitespace in message bytes

```
# detached cat capture; send spaced message verbatim
harness -> detached cat capture -> tty-watch send session-N "  spaced  " -> capture.bin == "  spaced  "
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "send-preserves-whitespace"
	req.SendMessage = "  spaced  "
	return nil
}
```