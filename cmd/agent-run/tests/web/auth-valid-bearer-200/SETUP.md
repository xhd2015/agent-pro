# Scenario

**Feature**: web health with valid Bearer token returns 200

```
agent-run web --token test → GET health Bearer test → 200
```

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).
- Web server started with `--token test`.

## Steps

1. Start `agent-run web --token test --port 0 --no-open`.
2. `GET /api/agent-run/health` with `Authorization: Bearer test`.

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
	req.HTTPAuth = "test"
	startWebServer(t, req)
	return nil
}
```