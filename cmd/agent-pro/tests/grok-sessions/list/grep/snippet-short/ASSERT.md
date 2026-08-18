## Expected Output

```
---
version: 3
---
SESSION ID                              KIND   LAST ACTIVE   TITLE                                        MSGS  CWD
01900021-aaaa-7aaa-aaaa-aaaaaaaaaaaa    main   30m ago       Ship GREP_SNIP_SHORT_TOKEN quickly              0  /tmp/grep-snippet-short
  summary\.json:1:title: Ship GREP_SNIP_SHORT_TOKEN quickly
```

## Expected

- One matching session: `01900021-aaaa-7aaa-aaaa-aaaaaaaaaaaa`.
- Hit on `summary.json:1:title:` with full title text in the snippet.
- `Hits[0].Snippet` equals the full short title (no truncation).
- Snippet does **not** contain `...` (no false windowing ellipsis).
- Snippet rune count is far below 1024.
- `MatchStart`/`MatchLen` span `GREP_SNIP_SHORT_TOKEN` inside the snippet.
- No ANSI escapes.

## Errors

- None.

```go
import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if len(resp.Sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(resp.Sessions))
	}
	if resp.Sessions[0].ID != "01900021-aaaa-7aaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("session id = %q, want 01900021-aaaa-7aaa-aaaa-aaaaaaaaaaaa", resp.Sessions[0].ID)
	}
	if len(resp.Matches) != 1 || len(resp.Matches[0].Hits) < 1 {
		t.Fatalf("want ≥1 hit, got matches=%d", len(resp.Matches))
	}

	const (
		title  = "Ship GREP_SNIP_SHORT_TOKEN quickly"
		needle = "GREP_SNIP_SHORT_TOKEN"
	)
	h := resp.Matches[0].Hits[0]
	if h.File != "summary.json" || h.Part != "title" {
		t.Fatalf("hit = %s:%d:%s, want summary.json:1:title", h.File, h.Line, h.Part)
	}
	if h.Snippet != title {
		t.Fatalf("snippet = %q, want full short title %q", h.Snippet, title)
	}
	if strings.Contains(h.Snippet, "...") {
		t.Fatalf("short snippet must not introduce windowing ellipsis: %q", h.Snippet)
	}
	rc := utf8.RuneCountInString(h.Snippet)
	if rc > 1024 {
		t.Fatalf("short snippet rune count = %d, unexpectedly > 1024", rc)
	}
	if h.MatchStart < 0 || h.MatchStart+h.MatchLen > len(h.Snippet) {
		t.Fatalf("MatchStart/MatchLen out of range: start=%d len=%d", h.MatchStart, h.MatchLen)
	}
	got := h.Snippet[h.MatchStart : h.MatchStart+h.MatchLen]
	if got != needle {
		t.Fatalf("match span = %q, want %q", got, needle)
	}

	assertNotContains(t, resp.Output, "\x1b[")
	// Full formatted output: short hit line has full title, no ellipsis in snippet portion.
	assert.Output(t, resp.Output, `---
version: 3
---
SESSION ID                              KIND   LAST ACTIVE   TITLE                                        MSGS  CWD
01900021-aaaa-7aaa-aaaa-aaaaaaaaaaaa    main   30m ago       Ship GREP_SNIP_SHORT_TOKEN quickly              0  /tmp/grep-snippet-short
  summary\.json:1:title: Ship GREP_SNIP_SHORT_TOKEN quickly`)
}
```
