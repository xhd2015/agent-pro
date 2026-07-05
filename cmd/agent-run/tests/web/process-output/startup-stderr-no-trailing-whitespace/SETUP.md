# Scenario

**Bug**: omit `--token` startup stderr must not end with trailing whitespace

```
agent-run web (no --token) -> stderr: listen URL line ends with \n only (no trailing space/tab)
```

## Preconditions

- `WebTokenMode=omit` exercises the open-API startup path.
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