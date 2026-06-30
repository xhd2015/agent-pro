# Scenario

**Bug**: web without `--token` warns users to configure API protection

```
agent-run web --port 0 (no --token) → stderr warns about --token / --token auto
```

## Preconditions

- Same startup as `no-token-health-200` (open API mode).

## Steps

1. Start `agent-run web --port 0 --no-open` without `--token`.
2. Capture startup stderr via `req.WebServerStderr` (no HTTP probe required for warning text).

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