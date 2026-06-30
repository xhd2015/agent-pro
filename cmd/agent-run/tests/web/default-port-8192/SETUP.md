# Scenario

**Feature**: `web` without `--port` binds default port 8192

```
agent-run web --token test (no --port) → 127.0.0.1:8192, GET health → 200
```

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).
- Port 8192 must be free for the test duration.

## Steps

1. Start `agent-run web --token test --no-open` without `--port` (default 8192).
2. Probe `127.0.0.1:8192` or parse stderr for listen on 8192.
3. `GET /api/agent-run/health` with valid Bearer.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "web"
	req.WebTokenMode = "explicit"
	req.WebToken = "test"
	req.WebPort = -1 // sentinel: omit --port flag, expect default 8192
	req.HTTPAuth = "test"
	startWebServer(t, req)
	return nil
}
```