## Expected

- `attachStdoutWriter` calls `normalizeTTYOutput` before writing when `rawTTY` is true.
- LF-only screen snapshot text must not pass through unchanged on interactive stdout.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.SourceCheckOK {
		t.Fatalf("source check failed: %s", resp.SourceCheckNote)
	}
	if !resp.AttachStdoutWriterNormalizesRawTTY {
		t.Fatalf("attachStdoutWriter must normalize LF-only output on raw TTY: %s", resp.SourceCheckNote)
	}
}
```