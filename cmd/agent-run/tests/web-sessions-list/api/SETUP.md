# Scenario

**Feature**: HTTP GET `/api/agent-run/sessions` contracts (pagination, q, status, counts)

```
# multi-step HTTP against sessions list API
agent-run web -> GET /api/agent-run/sessions[?limit&offset&q&status]
  -> JSON sessions + pagination + counts
```

## Preconditions

- Mode is `http` (no Playwright required).
- Explicit bearer token by default.
- Leaves seed flat session metas, start web, and populate `HTTPSteps`.

## Steps

1. Set `req.Mode = "http"`.
2. Default explicit token for health + API auth.
3. Leaves choose seed fixture and GET query strings.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "http"
	if req.WebTokenMode == "" {
		req.WebTokenMode = "explicit"
	}
	if req.WebTokenMode == "explicit" && req.Token == "" {
		req.Token = "test-token"
	}
	return nil
}
```
