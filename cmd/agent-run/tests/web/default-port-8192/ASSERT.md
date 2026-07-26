---
label: e2e
---

## Expected

- Server listens on default port **8192** when `--port` is omitted.
- Assert checks at least one of:
  - `req.WebBaseURL` contains `:8192`, or
  - stderr mentions `8192`, or
  - TCP connect to `127.0.0.1:8192` succeeds.
- Health returns HTTP 200 with valid Bearer.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	on8192 := strings.Contains(req.WebBaseURL, ":8192") ||
		portOpen("127.0.0.1", 8192)
	if !on8192 {
		t.Fatalf("expected server on default port 8192, WebBaseURL=%q", req.WebBaseURL)
	}
	if resp.HTTPStatus != 200 {
		t.Fatalf("expected HTTP 200 on default port, got %d body=%q", resp.HTTPStatus, resp.HTTPBody)
	}
}
```