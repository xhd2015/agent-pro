# Scenario

**Feature**: backend HTTP API exposes tty runners and terminal availability

```
agent-run web -> /api/agent-run/runners -> runner catalog includes tty runners
stored session + tty registry -> /terminal -> availability JSON
```

## Preconditions

- Web server is running with Bearer token auth.
- Terminal status endpoint is under `/api/agent-run/sessions/{runner}/{session_id}/terminal`.

## Steps

1. Descendant setup chooses runner/session fixture.
2. `Run` sends authenticated HTTP request to the target backend API endpoint.
3. Leaf `Assert` validates JSON and status code.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "http"
	req.HTTPAuth = req.WebToken
	if req.HTTPMethod == "" {
		req.HTTPMethod = "GET"
	}
	return nil
}
```
