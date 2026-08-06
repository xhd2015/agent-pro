## Expected

- No error.
- List length 2; each has 1 prompt (a1 and b1) and OmittedAfter=2.
- Output contains both session ids.
- Output contains a1 and b1; not a2/a3/b2/b3.
- `(...2 omitted...)` appears (at least twice, once per session block).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertListLen(t, resp.List, 2)
	for i, sp := range resp.List {
		if len(sp.UserPrompts) != 1 {
			t.Fatalf("session[%d] prompts=%d want 1", i, len(sp.UserPrompts))
		}
		if sp.OmittedAfter != 2 {
			t.Fatalf("session[%d] OmittedAfter=%d want 2", i, sp.OmittedAfter)
		}
	}
	assertPromptText(t, &resp.List[0], 0, "a1")
	assertPromptText(t, &resp.List[1], 0, "b1")

	out := resp.Output
	assertContains(t, out, idFilterGrepA)
	assertContains(t, out, idFilterGrepB)
	assertContains(t, out, "a1")
	assertContains(t, out, "b1")
	for _, drop := range []string{"a2", "a3", "b2", "b3"} {
		if strings.Contains(out, drop) {
			t.Fatalf("output contains dropped %q:\n%s", drop, out)
		}
	}
	if c := strings.Count(out, omissionMarker(2)); c < 2 {
		t.Fatalf("want >=2 omission markers, got %d:\n%s", c, out)
	}
}
```
