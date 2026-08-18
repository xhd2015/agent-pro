## Expected

- `IsGrokRunner` is true only for exact trimmed `grok` and `grok-tty`.
- Uppercase, prefix, and other runners are false.

```go
import (
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	if len(resp.IsGrok) != len(req.RunnerCases) {
		t.Fatalf("got %d IsGrok results, want %d", len(resp.IsGrok), len(req.RunnerCases))
	}
	for i, c := range req.RunnerCases {
		if resp.IsGrok[i] != c.Want {
			t.Errorf("case %q IsGrokRunner(%q)=%v, want %v", c.Name, c.Runner, resp.IsGrok[i], c.Want)
		}
	}
}
```
