# Scenario

**Feature**: HTTP API contracts for selected workspace, MRU, fs list, sessions

```
# multi-step HTTP against agent-run web APIs
agent-run web -> GET status | PUT workspace | GET fs/list | POST sessions
  -> JSON status / config.json / session meta.workspace
```

## Preconditions

- Mode is `http` (no Playwright required).
- Explicit bearer token by default.
- Leaves start web and populate `HTTPSteps`.

## Steps

1. Set `req.Mode = "http"`.
2. Default explicit token for health + API auth.
3. Leaves choose fixture paths and multi-step sequences.

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
