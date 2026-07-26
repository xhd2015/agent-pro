---
label: e2e
---

## Expected

- HTTP status 401 when Bearer token does not match `--token test`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if resp.HTTPStatus != 401 {
		t.Fatalf("expected HTTP 401 for wrong Bearer, got %d body=%q", resp.HTTPStatus, resp.HTTPBody)
	}
}
```