---
label: e2e
---

## Expected

- HTTP 200.
- JSON has `available: true`, same `runner`, and same `session_id`.
- Body does not leak registry filesystem paths.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertHTTPStatus(t, resp, 200)
	obj := decodeJSONBody(t, resp.HTTPBody)
	if !boolField(obj, "available") {
		t.Fatalf("terminal should be available: %s", resp.HTTPBody)
	}
	if stringField(obj, "runner") != req.Runner || stringField(obj, "session_id") != req.SessionID {
		t.Fatalf("wrong terminal identity: %s", resp.HTTPBody)
	}
	requireNoPathLeak(t, resp.HTTPBody)
}
```
