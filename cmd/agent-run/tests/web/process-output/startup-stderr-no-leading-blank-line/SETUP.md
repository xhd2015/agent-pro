# Scenario

**Bug**: omit `--token` startup stderr must not begin with a blank line

```
agent-run web (no --token) -> stderr: first bytes are warning text (no leading \n)
```

## Preconditions

- `WebTokenMode=omit` exercises the open-API startup path.
- No HTTP probe required beyond optional health check in `Run`.

## Steps

1. Start `agent-run web --port 0` without `--token`.
2. `Run` GET health (open API).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
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