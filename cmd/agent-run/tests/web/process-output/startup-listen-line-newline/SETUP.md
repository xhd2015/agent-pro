# Scenario

**Bug**: listen URL on stderr must end with a newline when `--token` is omitted

```
agent-run web (no --token) -> stderr: warning line + listening at URL line (terminated)
```

## Preconditions

- `WebTokenMode=omit` so `listenEndsWithNewline` is exercised.
- No HTTP probe required beyond optional health check in `Run`.

## Steps

1. Start `agent-run web --port 0` without `--token`.
2. `Run` GET health (open API).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "web"
	req.WebTokenMode = "omit"
	req.WebPort = 0
	startWebServer(t, req)
	req.HTTPMethod = "GET"
	req.HTTPAuth = ""
	req.HTTPPath = "/api/agent-run/health"
	return nil
}
```