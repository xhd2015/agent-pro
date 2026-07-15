## Expected

- `IsTransientIndexError` returns true for the index-write variant without "new".
- `resp.Transient` is true.

## Errors

- None from `Run`.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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
		t.Fatalf("unable to write index file variant must be transient: %q", req.ClassifyOutput)
	}
}
```
