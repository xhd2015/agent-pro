# Scenario

**Bug**: web without `--token` must allow unauthenticated API access

```
agent-run web --port 0 (no --token) → GET /api/agent-run/health (no Authorization) → 200
```

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).
- Web started with **no** `--token` flag (`req.WebTokenMode = "omit"`).

## Steps

1. Start `agent-run web --port 0 --no-open` without `--token`.
2. `GET /api/agent-run/health` without `Authorization`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "web"
	req.WebTokenMode = "omit"
	req.WebPort = 0
	req.HTTPAuth = ""
	startWebServer(t, req)
	return nil
}
```