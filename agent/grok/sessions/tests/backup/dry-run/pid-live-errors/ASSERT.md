## Expected

- `Backup` returns an error (live process / pid / running / busy wording).
- Result is nil.
- OutDir has no payload/manifest and is not created.

## Errors

- Mentions live process, pid, or running/busy session.

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
	has := strings.Contains(msg, "pid") ||
		strings.Contains(msg, "live") ||
		strings.Contains(msg, "running") ||
		strings.Contains(msg, "process") ||
		strings.Contains(msg, "busy")
	if !has {
		t.Fatalf("error %q should mention live pid/process", resp.Err)
	}
	assertNoPayloadUnder(t, req.OutDir)
	assertPathMissing(t, req.OutDir)
}
```
