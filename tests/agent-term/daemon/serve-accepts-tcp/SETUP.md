# Scenario

**Feature**: serve binds TCP and accepts HTTP

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "serve-tcp"
	return nil
}
```