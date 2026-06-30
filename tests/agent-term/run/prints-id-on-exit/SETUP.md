# Scenario

**Feature**: run prints only session id after command exits

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "run-exit-id"
	req.StartDaemon = true
	return nil
}
```