## Expected

- Find is not-found for the former UUID.
- `index/by-runner-session/` is absent or has no UUID cache files after clear
  (implementer may delete the mapping dir or the whole `index/`).

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
	snap := resp.CacheAfter
	if snap.ByRunnerExists && len(snap.UUIDFiles) > 0 {
		t.Fatalf("after ClearAllSessions expected empty/absent by-runner-session mapping, have files %v", snap.UUIDFiles)
	}
}
```
