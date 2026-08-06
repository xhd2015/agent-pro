## Expected

- `Backup` returns an error (file-active / busy / active wording).
- Result is nil.
- OutDir has no `manifest.json` / `payload/` and is not created as a backup tree.

## Errors

- Mentions that the session is active / file-active / in use (or busy).

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
	if !strings.Contains(msg, "active") && !strings.Contains(msg, "busy") && !strings.Contains(msg, "in use") {
		t.Fatalf("error %q should mention active/busy session", resp.Err)
	}
	assertNoPayloadUnder(t, req.OutDir)
	assertPathMissing(t, req.OutDir)
}
```
