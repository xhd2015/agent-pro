# Scenario

**Feature**: serve logs listen address to stderr on startup

```
# daemon observability
agent-term serve --listen ADDR -> stderr: listening on ADDR
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "serve-logs-startup"
	return nil
}
```