# Scenario

**Feature**: web health without Bearer token returns 401

```
agent-run web --token test → GET /api/agent-run/health (no Authorization) → 401
```

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).
- Web server started with `--token test`.

## Steps

1. Start `agent-run web --token test --port 0 --no-open`.
2. `GET /api/agent-run/health` without `Authorization` header.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "web"
	req.WebTokenMode = "explicit"
	req.WebToken = "test"
	req.WebPort = 0
	req.HTTPAuth = ""
	startWebServer(t, req)
	return nil
}
```