## Expected

- `Backup` returns an error (exists / already).
- Result is nil.
- Pre-existing archive bytes unchanged.
- OutDir has no payload/manifest.

## Errors

- Mentions that the archive path exists / already exists.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertError(t, resp)
	if resp.Result != nil {
		t.Fatalf("expected nil Result, got %+v", resp.Result)
	}
	msg := strings.ToLower(resp.Err.Error())
	if !strings.Contains(msg, "exist") && !strings.Contains(msg, "already") {
		t.Fatalf("error %q should mention existing archive path", resp.Err)
	}
	got := readFileString(t, req.ArchivePath)
	if got != "pre-existing-archive-bytes\n" {
		t.Fatalf("pre-existing archive was modified: %q", got)
	}
	assertNoPayloadUnder(t, req.OutDir)
}
```
