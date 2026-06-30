# Scenario

**Feature**: `web --port 0` starts and health endpoint responds

```
agent-run web --port 0 --token test → OS-assigned port, GET health Bearer test → 200
```

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).

## Steps

1. Start `agent-run web --token test --port 0 --no-open`.
2. `GET /api/agent-run/health` with valid Bearer after server is ready.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "web"
	req.WebTokenMode = "explicit"
	req.WebToken = "test"
	req.WebPort = 0
	req.HTTPAuth = "test"
	startWebServer(t, req)
	return nil
}
```