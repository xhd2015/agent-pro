# Scenario

**Feature**: run creates session, attaches, prints id on exit

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.StartDaemon = true
	return nil
}
```