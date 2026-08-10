## Expected

- `PublishCount` == 0.
- Capture has no requests.
- Run `err` is nil.

## Side Effects

- None (send path must not touch the bus).

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if resp.PublishCount != 0 {
		t.Fatalf("send path must not publish; PublishCount=%d", resp.PublishCount)
	}
	if req.Capture.Len() != 0 {
		t.Fatalf("send path must not record HTTP; got %d", req.Capture.Len())
	}
}
```
