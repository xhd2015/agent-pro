# Scenario

**Feature**: registry session stays live while headless parent is blocked waiting

```
# background headless after session-id; list while parent still running
harness pipe -> tty-watch run --headless sleep 120 -> list shows sleep 120
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-headless-session-live-while-waiting"
	req.RunCommand = []string{"sleep", "120"}
	return nil
}
```