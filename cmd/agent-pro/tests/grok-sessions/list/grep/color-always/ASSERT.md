## Expected Output

Hit line (under the session row) should colorize segments similarly to:

```
  <ansi-color #35>summary.json</ansi-color>:<ansi-color green>1</ansi-color>:<ansi-color green>title</ansi-color>: Enable <ansi-color bold red>GREP_COLOR_TOKEN</ansi-color> highlighting
```

## Expected

- Session is listed.
- Output contains ANSI CSI sequences (`\x1b[`).
- Filename portion uses magenta, line/part use green, match uses bold red.
- MatchStart/MatchLen on the hit enable correct match span coloring.

## Errors

- None.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if len(resp.Sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(resp.Sessions))
	}
	if !strings.Contains(resp.Output, "\x1b[") {
		t.Fatalf("expected ANSI escapes with Color=always, got:\n%s", resp.Output)
	}
	// Magenta filename
	assertContains(t, resp.Output, "\x1b[35m")
	// Green for line number and/or part
	assertContains(t, resp.Output, "\x1b[32m")
	// Bold and red for match highlight (may be combined as \x1b[1;31m or sequential)
	if !strings.Contains(resp.Output, "\x1b[1m") && !strings.Contains(resp.Output, "\x1b[1;") {
		t.Fatalf("expected bold SGR in colored output:\n%s", resp.Output)
	}
	if !strings.Contains(resp.Output, "\x1b[31m") && !strings.Contains(resp.Output, ";31m") {
		t.Fatalf("expected red SGR in colored output:\n%s", resp.Output)
	}
	assertContains(t, resp.Output, "GREP_COLOR_TOKEN")
	assertContains(t, resp.Output, "summary.json")

	// Prefer structured match offsets for implementer-friendly coloring.
	if len(resp.Matches) == 1 && len(resp.Matches[0].Hits) >= 1 {
		h := resp.Matches[0].Hits[0]
		if h.MatchLen <= 0 {
			t.Fatalf("MatchLen = %d, want > 0 for color span", h.MatchLen)
		}
		if h.MatchStart < 0 || h.MatchStart+h.MatchLen > len(h.Snippet) {
			t.Fatalf("MatchStart/MatchLen out of range: start=%d len=%d snippet=%q",
				h.MatchStart, h.MatchLen, h.Snippet)
		}
	}

	// Soft full-line color template check on the hit line only (table row may be plain).
	// Extract the first indented hit line for template match.
	var hitLine string
	for _, line := range strings.Split(resp.Output, "\n") {
		if strings.Contains(line, "summary.json") && strings.Contains(line, "title") {
			hitLine = line
			break
		}
	}
	if hitLine == "" {
		t.Fatalf("missing summary.json title hit line in:\n%s", resp.Output)
	}
	assert.Output(t, hitLine+"\n", `---
version: 2
---
  <ansi-color #35>summary.json</ansi-color>:<ansi-color green>1</ansi-color>:<ansi-color green>title</ansi-color>: Enable <ansi-color bold red>GREP_COLOR_TOKEN</ansi-color> highlighting
`)
}
```
