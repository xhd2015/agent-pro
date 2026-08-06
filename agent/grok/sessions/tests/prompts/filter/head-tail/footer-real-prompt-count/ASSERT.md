## Expected

- No error.
- Output contains omission marker `(...3 omitted...)`.
- Footer reports **2** user messages (printed), e.g. `1 sessions, 2 user messages`.
- Footer must not claim 5 (pre-slice) or 3 (including virtual line).

```go
import (
	"regexp"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	out := resp.Output
	assertContains(t, out, omissionMarker(3))
	// Accept "1 session" or "1 sessions" wording variants; require 2 user messages.
	re := regexp.MustCompile(`(?i)1\s+sessions?,\s*2\s+user messages`)
	if !re.MatchString(out) {
		t.Fatalf("footer must count 2 printed user messages, got:\n%s", out)
	}
	bad := regexp.MustCompile(`(?i)\d+\s+sessions?,\s*(3|5)\s+user messages`)
	if bad.MatchString(out) {
		t.Fatalf("footer incorrectly counts virtual or pre-slice prompts:\n%s", out)
	}
}
```
