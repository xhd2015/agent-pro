## Expected
- Zip file is created (empty — no files added).
- No error — missing source directory is skipped gracefully.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	entries := readZipEntries(t, req.ZipPath)
	if len(entries) != 0 {
		t.Fatalf("expected empty zip, got %d entries: %v", len(entries), entries)
	}
}
```
