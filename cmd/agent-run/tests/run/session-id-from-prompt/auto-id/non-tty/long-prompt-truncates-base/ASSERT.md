## Expected

- Exit code 0.
- Session id matches auto-id shape.
- Base portion has at most 128 runes.
- Base is a prefix of the long `a…` slug (all `a`s after truncate/re-trim).

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)

	id := singleSessionID(t, req.Home, "fake-codex")
	base, ts, _, ok := splitAutoSessionID(id)
	if !ok {
		t.Fatalf("session id %q does not match auto-id shape", id)
	}
	if ts == "" {
		t.Fatalf("missing timestamp in id %q", id)
	}
	n := runeCount(base)
	if n > 128 {
		t.Fatalf("base slug rune count = %d, want ≤ 128 (base=%q id=%q)", n, base, id)
	}
	if n == 0 {
		t.Fatalf("empty base after truncate (id=%q)", id)
	}
	if strings.Trim(base, "a") != "" {
		t.Fatalf("expected base of only 'a' runes after truncate, got %q", base)
	}
	// Without truncation the base would be 200 a's; must be shorter.
	if n >= 200 {
		t.Fatalf("base was not truncated: rune count %d (id=%q)", n, id)
	}
}
```
