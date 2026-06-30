# Scenario

**Feature**: web health with wrong Bearer token returns 401

```
agent-run web --token test → GET health Bearer wrong-token → 401
```

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).
- Web server started with `--token test`.

## Steps

1. Start `agent-run web --token test --port 0 --no-open`.
2. `GET /api/agent-run/health` with `Authorization: Bearer wrong-token`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "web"
	req.WebTokenMode = "explicit"
	req.WebToken = "test"
	req.WebPort = 0
	req.HTTPAuth = "wrong-token"
	startWebServer(t, req)
	return nil
}
```