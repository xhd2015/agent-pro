# Scenario

**Feature**: detached run with `--session-id` writes registry and appears in list

```
# reserve custom id, detach, list shows id + command
harness PTY -> tty-watch run --session-id test-with-grok sleep 300 -> \x1d -> list
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-custom-registers"
	req.CustomSessionID = "test-with-grok"
	req.RunCommand = []string{"sleep", "300"}
	return nil
}
```