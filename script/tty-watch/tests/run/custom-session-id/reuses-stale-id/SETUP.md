# Scenario

**Feature**: stale registry for custom id is pruned and id reused on new run

```
# seed unreachable registry entry for custom id
harness seeds test-with-grok.json (dead listen addr)
# run reuses id and becomes live
harness PTY -> tty-watch run --session-id test-with-grok sleep 300 -> \x1d -> list
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-custom-reuses-stale"
	req.CustomSessionID = "test-with-grok"
	req.RunCommand = []string{"sleep", "300"}
	return nil
}
```