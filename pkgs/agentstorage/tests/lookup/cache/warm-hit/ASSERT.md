## Expected

- Second Find returns `warm-hit`.
- `CacheAfter.ByRunnerGen` equals `WarmGen` (unchanged across warm hit).

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
	if resp.Meta.SessionID != "warm-hit" {
		t.Fatalf("SessionID=%q, want warm-hit", resp.Meta.SessionID)
	}
	if resp.WarmGen == "" {
		t.Fatal("expected WarmGen after warm populate")
	}
	if resp.CacheAfter.ByRunnerGen != resp.WarmGen {
		t.Fatalf(".gen after warm hit %q != WarmGen %q", resp.CacheAfter.ByRunnerGen, resp.WarmGen)
	}
}
```
