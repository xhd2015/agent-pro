# Scenario

**Feature**: agent-term serve starts TCP daemon

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "serve-tcp"
	return nil
}
```