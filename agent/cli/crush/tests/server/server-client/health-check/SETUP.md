# Scenario

**Feature**: ensureServer probes the /v1/health endpoint and starts the server if needed

## Preconditions
- A crush server may or may not be running.
- `ensureServer` should probe `/v1/health` and start the server if needed.

## Steps
1. Set `ServerOperation` to `"health-check"` to probe the health endpoint.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ServerOperation = "health-check"
	return nil
}
```
