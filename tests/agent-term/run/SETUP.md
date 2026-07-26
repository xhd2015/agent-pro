# Scenario

**Feature**: run creates session, attaches, prints id on exit

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.StartDaemon = true
	return nil
}
```