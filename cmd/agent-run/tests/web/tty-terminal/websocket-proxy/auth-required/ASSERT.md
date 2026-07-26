---
label: e2e
---

## Expected

- Websocket handshake is rejected with HTTP 401 or 403.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.HTTPStatus != 401 && resp.HTTPStatus != 403 {
		t.Fatalf("unauthorized websocket status=%d want 401/403 error=%q output=%q", resp.HTTPStatus, resp.WSError, resp.WSOutput)
	}
}
```
