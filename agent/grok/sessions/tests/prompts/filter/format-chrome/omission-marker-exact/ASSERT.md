## Expected

- No error.
- Output contains the exact characters `(...3 omitted...)` as a full line.
- Not variants like `... 3 omitted` without parens, or `[3 omitted]`.

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
	want := "(...3 omitted...)"
	if !strings.Contains(resp.Output, want) {
		t.Fatalf("missing exact marker %q in:\n%s", want, resp.Output)
	}
	// full-line presence
	found := false
	for _, ln := range strings.Split(resp.Output, "\n") {
		// allow dim ANSI wrapping around the marker when color on; strip CSI
		plain := stripANSI(ln)
		if plain == want {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("marker not present as its own line (plain):\n%s", resp.Output)
	}
}

func stripANSI(s string) string {
	// minimal CSI strip for marker line checks
	var b strings.Builder
	in := false
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b {
			in = true
			continue
		}
		if in {
			if (s[i] >= 'a' && s[i] <= 'z') || (s[i] >= 'A' && s[i] <= 'Z') {
				in = false
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
```
