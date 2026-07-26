# Scenario

**Feature**: attach connects to existing session by id or name

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.StartDaemon = true
	return nil
}
```