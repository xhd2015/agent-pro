# Scenario

**Bug**: web API and session meta must expose the server working directory as `workspace`

```
GET /api/agent-run/status -> {home, workspace}
POST create session -> GET detail -> session.workspace matches server cwd
```

## Preconditions

- Web process `Dir` is the test temp dir (inherited `startWebServer` uses `req.TempDir`).
- `workspace` is read-only display metadata (no remote picker).

## Steps

1. Leaf starts web server with explicit token.
2. Leaf sets `HTTPPath` to status or session detail endpoint.
3. `Run` performs GET; `Assert` checks JSON `workspace` field.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.WebTokenMode == "" {
		req.WebTokenMode = "explicit"
	}
	if req.WebToken == "" {
		req.WebToken = "test"
	}
	return nil
}
```