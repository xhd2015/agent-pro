## Expected

- Find error is not-found for UUID Z.
- Cache has no file for Z (warm miss uses missing file = empty matches; no scan write).

```go
import (
	"fmt"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	want := fmt.Sprintf("session not found: no grok session with runner_session_id %s", req.QueryID)
	assertExactErr(t, resp.Err, want)
	uuidZ := "99999999-9999-9999-9999-999999999999"
	if cacheHasUUID(resp.CacheAfter, uuidZ) {
		t.Fatalf("warm known-miss must not create cache file for %s; have %v", uuidZ, resp.CacheAfter.UUIDFiles)
	}
}
```
