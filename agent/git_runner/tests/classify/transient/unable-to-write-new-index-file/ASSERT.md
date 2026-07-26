## Expected

- `IsTransientIndexError` returns true for the exact production message.
- `resp.Transient` equals `req.WantTransient` (true).

## Errors

- None from `Run`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil Response")
	}
	if resp.Transient != req.WantTransient {
		t.Fatalf("IsTransientIndexError(%q) = %v, want %v", req.ClassifyOutput, resp.Transient, req.WantTransient)
	}
	if !resp.Transient {
		t.Fatalf("production message must be transient: %q", req.ClassifyOutput)
	}
}
```
