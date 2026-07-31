## Expected

- Returned TTYs include `/dev/ttys148` (ancestor) and `/dev/ttys200` (descendant).
- Returned TTYs do **not** include `??`, blank, or unrelated `/dev/ttys999`.
- Each included TTY appears at most once (dedup OK if implementer chooses).

## Errors

- None from `Run`.

## Exit Code

- N/A (library)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	got := map[string]int{}
	for _, tty := range resp.TTYs {
		got[tty]++
		if strings.TrimSpace(tty) == "" || tty == "??" {
			t.Fatalf("CollectTTYsFromTree must skip blank/??, got %q in %v", tty, resp.TTYs)
		}
	}
	for _, want := range []string{"/dev/ttys148", "/dev/ttys200"} {
		if got[want] == 0 {
			t.Fatalf("missing TTY %q in %v", want, resp.TTYs)
		}
		if got[want] > 1 {
			t.Fatalf("TTY %q duplicated in %v", want, resp.TTYs)
		}
	}
	if got["/dev/ttys999"] != 0 {
		t.Fatalf("unrelated process TTY must not appear: %v", resp.TTYs)
	}
}
```
