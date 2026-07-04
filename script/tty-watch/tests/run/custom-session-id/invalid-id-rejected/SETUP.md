# Scenario

**Feature**: invalid custom session ids are rejected before starting a session

```
# leading dot violates validation pattern
tty-watch run --session-id .bad sleep 1 -> exit 1
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-custom-invalid-id"
	req.CustomSessionID = ".bad"
	req.RunCommand = []string{"sleep", "1"}
	return nil
}
```