# Scenario

**Feature**: serve binds TCP and accepts HTTP

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "serve-tcp"
	return nil
}
```