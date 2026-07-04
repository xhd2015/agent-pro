# Scenario

**Feature**: `list` shows custom ids alongside auto-reserved `session-N` ids

```
# detached custom id + detached auto session-N coexist
harness PTY -> run --session-id test-with-grok sleep 300 -> \x1d
harness PTY -> run sleep 300 -> \x1d -> list both ids
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-custom-list-mixed"
	req.CustomSessionID = "test-with-grok"
	req.RunCommand = []string{"sleep", "300"}
	return nil
}
```