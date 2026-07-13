# Scenario

**Feature**: server-side SPA static fallback and path parsing via HTTP

```
# non-API paths -> index.html shell; optional session bootstrap; API stays API
Browser -> GET / | /sessions/:id | wrong shapes -> registerStatic -> HTML
Browser -> GET /api/agent-run/... -> API mux (not SPA HTML success body)
```

## Preconditions

- Leaf starts `agent-run web` after optional session seed.
- Mode is `http` (no Playwright required).

## Steps

1. Set `req.Mode = "http"`.
2. Default explicit token so health wait works consistently.
3. Leaves choose path(s) and assert HTML vs API contracts.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "http"
	if req.WebTokenMode == "" {
		req.WebTokenMode = "explicit"
	}
	if req.WebTokenMode == "explicit" && req.Token == "" {
		req.Token = "test-token"
	}
	return nil
}
```
