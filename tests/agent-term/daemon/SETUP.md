# Scenario

**Feature**: agent-term serve starts TCP daemon

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "serve-tcp"
	return nil
}
```