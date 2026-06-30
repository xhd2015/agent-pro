# Scenario

**Feature**: `--token auto` generates a bearer token and enforces auth

```
agent-run web --token auto --port 0 → stderr prints token → health without auth → 401
```

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).

## Steps

1. Start `agent-run web --token auto --port 0 --no-open`.
2. `Run` performs an initial health GET without Bearer (leaf default); `Assert` verifies stderr, 401/200 with parsed token, and `auth.token` persistence.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "web"
	req.WebTokenMode = "auto"
	req.WebPort = 0
	req.HTTPAuth = ""
	startWebServer(t, req)
	return nil
}
```