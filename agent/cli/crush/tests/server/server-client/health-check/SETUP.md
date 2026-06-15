## Preconditions
- A crush server may or may not be running.
- `ensureServer` should probe `/v1/health` and start the server if needed.

## Steps
1. Set `ServerOperation` to `"health-check"` to probe the health endpoint.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ServerOperation = "health-check"
	return nil
}
```
