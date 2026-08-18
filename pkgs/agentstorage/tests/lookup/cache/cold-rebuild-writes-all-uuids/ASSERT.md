## Expected

- Find returns `cold-a`.
- After call: `index/by-runner-session/` exists; `.gen` non-empty.
- Cache files exist for both queried UUID A and sibling UUID B.

## Side Effects

- Cold miss populates the lazy index from a full scan (not only the queried key).

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.Err != nil {
		t.Fatalf("Find error: %v", resp.Err)
	}
	if resp.Meta.SessionID != "cold-a" {
		t.Fatalf("SessionID=%q, want cold-a", resp.Meta.SessionID)
	}
	snap := resp.CacheAfter
	if !snap.ByRunnerExists {
		t.Fatal("expected index/by-runner-session/ after cold Find")
	}
	if snap.ByRunnerGen == "" {
		t.Fatal("expected non-empty .gen after cold rebuild")
	}
	uuidA := "11111111-1111-1111-1111-111111111111"
	uuidB := "22222222-2222-2222-2222-222222222222"
	if !cacheHasUUID(snap, uuidA) {
		t.Fatalf("missing cache file for queried UUID %s; have %v", uuidA, snap.UUIDFiles)
	}
	if !cacheHasUUID(snap, uuidB) {
		t.Fatalf("missing cache file for sibling UUID %s; have %v", uuidB, snap.UUIDFiles)
	}
}
```
