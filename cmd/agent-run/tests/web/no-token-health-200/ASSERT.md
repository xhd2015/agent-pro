---
label: e2e
---

## Expected

- HTTP status **200** on `/api/agent-run/health` with no `Authorization` header.

## Side Effects

- `auth.token` must **not** be created under `AGENT_RUN_HOME` for open (no `--token`) mode.

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.HTTPStatus != 200 {
		t.Fatalf("expected HTTP 200, got %d body=%q", resp.HTTPStatus, resp.HTTPBody)
	}
	authPath := filepath.Join(req.Home, "auth.token")
	if _, err := os.Stat(authPath); err == nil {
		t.Fatalf("auth.token must not exist in no-token mode: %s", authPath)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat auth.token: %v", err)
	}
}
```