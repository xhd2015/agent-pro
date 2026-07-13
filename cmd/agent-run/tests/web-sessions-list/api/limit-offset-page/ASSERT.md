## Expected

- All three GETs **200**.
- **page0**: length 2; ids `sess-epsilon`, `sess-delta`; `total=5`; `limit=2`; `offset=0`; `has_more=true`.
- **page1**: length 2; ids `sess-gamma`, `sess-beta`; no overlap with page0.
- **page2**: length 1; id `sess-alpha`; `has_more=false`.

## Errors

- Pre-impl: full list returned ignoring limit (RED).

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}

	p0 := requireOK200(t, resp, "page0")
	s0 := sessionsFromBody(t, p0.Body)
	if len(s0) != 2 {
		t.Fatalf("page0 length: got %d want 2 ids=%v", len(s0), sessionIDs(s0))
	}
	ids0 := sessionIDs(s0)
	if ids0[0] != "sess-epsilon" || ids0[1] != "sess-delta" {
		t.Fatalf("page0 order: got %v want [sess-epsilon sess-delta]", ids0)
	}
	m0 := parseJSONMap(t, p0.Body)
	total, ok := jsonFloat(m0, "total")
	if !ok {
		t.Fatal("page0 missing total")
	}
	if int(total) != 5 {
		t.Fatalf("page0 total: got %v want 5", total)
	}
	lim, ok := jsonFloat(m0, "limit")
	if !ok || int(lim) != 2 {
		t.Fatalf("page0 limit: got %v want 2 (present=%v)", lim, ok)
	}
	off, ok := jsonFloat(m0, "offset")
	if !ok || int(off) != 0 {
		t.Fatalf("page0 offset: got %v want 0 (present=%v)", off, ok)
	}
	hasMore, ok := jsonBool(m0, "has_more")
	if !ok {
		t.Fatal("page0 missing has_more")
	}
	if !hasMore {
		t.Fatal("page0 has_more: got false want true")
	}

	p1 := requireOK200(t, resp, "page1")
	s1 := sessionsFromBody(t, p1.Body)
	if len(s1) != 2 {
		t.Fatalf("page1 length: got %d want 2 ids=%v", len(s1), sessionIDs(s1))
	}
	ids1 := sessionIDs(s1)
	if ids1[0] != "sess-gamma" || ids1[1] != "sess-beta" {
		t.Fatalf("page1 order: got %v want [sess-gamma sess-beta]", ids1)
	}
	// no overlap
	set0 := map[string]bool{ids0[0]: true, ids0[1]: true}
	for _, id := range ids1 {
		if set0[id] {
			t.Fatalf("page1 id %q overlaps page0 %v", id, ids0)
		}
	}

	p2 := requireOK200(t, resp, "page2")
	s2 := sessionsFromBody(t, p2.Body)
	if len(s2) != 1 {
		t.Fatalf("page2 length: got %d want 1 ids=%v", len(s2), sessionIDs(s2))
	}
	if sessionIDs(s2)[0] != "sess-alpha" {
		t.Fatalf("page2 id: got %q want sess-alpha", sessionIDs(s2)[0])
	}
	m2 := parseJSONMap(t, p2.Body)
	hasMore2, ok := jsonBool(m2, "has_more")
	if !ok {
		t.Fatal("page2 missing has_more")
	}
	if hasMore2 {
		t.Fatal("page2 has_more: got true want false")
	}
}
```
