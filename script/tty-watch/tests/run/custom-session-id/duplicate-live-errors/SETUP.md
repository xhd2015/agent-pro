# Scenario

**Feature**: second `run --session-id` with same live id fails

```
# first detached custom session stays live
harness PTY -> tty-watch run --session-id test-with-grok sleep 300 -> \x1d
# second run with same id errors
tty-watch run --session-id test-with-grok sleep 300 -> exit 1
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-custom-duplicate-live"
	req.CustomSessionID = "test-with-grok"
	req.RunCommand = []string{"sleep", "300"}
	return nil
}
```