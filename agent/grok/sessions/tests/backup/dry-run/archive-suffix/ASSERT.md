## Expected

- `Backup` returns an error about archive suffix / `.tar.gz`.
- Result is nil.
- OutDir has no payload/manifest and is not created.
- Invalid archive path was not written.

## Errors

- Mentions `.tar.gz` or archive suffix/extension.

```go
import (
	"os"
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
	if !strings.Contains(msg, ".tar.gz") && !strings.Contains(msg, "tar.gz") && !strings.Contains(msg, "suffix") && !strings.Contains(msg, "extension") {
		t.Fatalf("error %q should mention .tar.gz suffix", resp.Err)
	}
	assertNoPayloadUnder(t, req.OutDir)
	assertPathMissing(t, req.OutDir)
	if _, err := os.Stat(req.ArchivePath); err == nil {
		t.Fatalf("archive path should not be written: %s", req.ArchivePath)
	}
}
```
