# Scenario

**Feature**: session detail page WorkspacePath — meta.workspace source

```
# explicit token + flat session seed
seed sessions/<id>/meta.json(workspace) -> web --token test-token
  -> GET /sessions/<id> -> session header WorkspacePath
```

## Preconditions

- Explicit Bearer token; Playwright seeds `localStorage['agent-run-token']`.
- Flat session layout: `AGENT_RUN_HOME/sessions/<session_id>/`.
- Client route: `/sessions/:sessionId`.

## Steps

1. Set explicit token mode defaults for session surface.
2. Leaves seed workspace + start web + open session URL.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.WebTokenMode = "explicit"
	if req.Token == "" {
		req.Token = "test-token"
	}
	return nil
}
```
