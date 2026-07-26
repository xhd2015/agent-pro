---
label: e2e
---

## Expected

- HTTP status 401 on `/api/agent-run/health` when no `Authorization` header is sent.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if resp.HTTPStatus != 401 {
		t.Fatalf("expected HTTP 401, got %d body=%q", resp.HTTPStatus, resp.HTTPBody)
	}
}
```