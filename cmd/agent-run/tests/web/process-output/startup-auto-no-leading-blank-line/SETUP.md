# Scenario

**Bug**: `--token auto` startup stderr must not begin with a blank line

```
agent-run web --token auto -> stderr: first bytes are token line (no leading \n)
```

## Preconditions

- `WebTokenMode=auto` exercises the generated-token startup path.
- No HTTP probe required beyond optional health check in `Run`.

## Steps

1. Start `agent-run web --token auto --port 0 --no-open`.
2. `Run` GET health without Bearer (leaf default).

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