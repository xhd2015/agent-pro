---
label: e2e
---

## Expected

- HTTP **200**.
- Body contains `#root` SPA mount.
- Body does **not** include `agent-run-session-bootstrap` (session missing → no inject).

## Side Effects

- Background web process cleaned up.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.HTTPStatus != 200 {
		t.Fatalf("expected 200, got %d body=%q", resp.HTTPStatus, resp.HTTPBody)
	}
	if !htmlHasRootMount(resp.HTTPBody) {
		t.Fatalf("missing #root in session-path HTML")
	}
	if htmlHasSessionBootstrap(resp.HTTPBody) {
		t.Fatalf("unknown session must not inject agent-run-session-bootstrap; body=%q", resp.HTTPBody)
	}
}
```
