# Scenario

**Feature**: status endpoint reports `home` and current `workspace` (cwd)

```
agent-run web -> GET /api/agent-run/status -> JSON workspace non-empty
```

## Preconditions

- Route `/api/agent-run/status` registered on web mux.
- Bearer token `test` required.

## Steps

1. Start web server.
2. `Run` GETs `/api/agent-run/status` with Bearer.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "web"
	req.WebTokenMode = "explicit"
	req.WebToken = "test"
	req.WebPort = 0
	startWebServer(t, req)
	req.HTTPMethod = "GET"
	req.HTTPAuth = req.WebToken
	req.HTTPPath = "/api/agent-run/status"
	return nil
}
```