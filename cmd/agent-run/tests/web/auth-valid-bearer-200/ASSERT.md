---
label: e2e
---

## Expected

- HTTP status 200 on `/api/agent-run/health` with valid `Authorization: Bearer test`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if resp.HTTPStatus != 200 {
		t.Fatalf("expected HTTP 200, got %d body=%q", resp.HTTPStatus, resp.HTTPBody)
	}
}
```