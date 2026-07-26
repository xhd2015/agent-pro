## Expected

- `IsTransientIndexError` returns false for empty commit message.
- `resp.Transient` is false.

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
	if resp.Transient {
		t.Fatalf("empty commit message must not be transient: %q", req.ClassifyOutput)
	}
}
```
