## Expected

- One matching session: `01900020-aaaa-7aaa-aaaa-aaaaaaaaaaaa`.
- One hit on `chat_history.jsonl:1:tool_result:` with pattern `SNIP_WIN_NEEDLE`.
- `Hits[0].Snippet` has **≤ 1024 runes** (`utf8.RuneCountInString`, not byte `len`).
- Snippet contains the exact match `SNIP_WIN_NEEDLE`.
- Snippet has leading and trailing ASCII ellipsis `...` (field was cut on both sides).
- Snippet has no raw newlines/tabs (whitespace collapsed to single spaces).
- Full unwindowed body is **not** dumped (e.g. full `600×'a'` prefix is absent).
- `MatchStart`/`MatchLen` are byte offsets into the final snippet spanning exactly
  `SNIP_WIN_NEEDLE`.
- No ANSI escapes (color never).

## Errors

- None.

```go
import (
	"strings"
	"testing"
	"unicode/utf8"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if len(resp.Sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(resp.Sessions))
	}
	if resp.Sessions[0].ID != "01900020-aaaa-7aaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("session id = %q, want 01900020-aaaa-7aaa-aaaa-aaaaaaaaaaaa", resp.Sessions[0].ID)
	}
	if len(resp.Matches) != 1 || len(resp.Matches[0].Hits) != 1 {
		t.Fatalf("want 1 match with 1 hit, got matches=%d hits=%v",
			len(resp.Matches), hitLens(resp))
	}

	h := resp.Matches[0].Hits[0]
	if h.File != "chat_history.jsonl" || h.Line != 1 || h.Part != "tool_result" {
		t.Fatalf("hit meta = %s:%d:%s, want chat_history.jsonl:1:tool_result", h.File, h.Line, h.Part)
	}

	const needle = "SNIP_WIN_NEEDLE"
	snippet := h.Snippet
	rc := utf8.RuneCountInString(snippet)
	if rc > 1024 {
		t.Fatalf("snippet rune count = %d, want ≤ 1024 (snippet=%q)", rc, truncateForLog(snippet, 80))
	}
	if !strings.Contains(snippet, needle) {
		t.Fatalf("snippet missing %q: %q", needle, truncateForLog(snippet, 120))
	}
	if !strings.HasPrefix(snippet, "...") {
		t.Fatalf("snippet missing leading ... ellipsis: %q", truncateForLog(snippet, 80))
	}
	if !strings.HasSuffix(snippet, "...") {
		t.Fatalf("snippet missing trailing ... ellipsis: %q", truncateForLog(snippet, 80))
	}
	if strings.ContainsAny(snippet, "\n\r\t") {
		t.Fatalf("snippet still contains raw whitespace runs (newlines/tabs): %q", truncateForLog(snippet, 80))
	}
	// Unwindowed body would include the full 600-rune prefix of a's.
	if strings.Contains(snippet, strings.Repeat("a", 600)) {
		t.Fatalf("snippet still contains full 600×'a' prefix (not windowed):\n%s", truncateForLog(snippet, 100))
	}
	if strings.Contains(snippet, strings.Repeat("b", 600)) {
		t.Fatalf("snippet still contains full 600×'b' suffix (not windowed):\n%s", truncateForLog(snippet, 100))
	}

	if h.MatchStart < 0 || h.MatchLen <= 0 || h.MatchStart+h.MatchLen > len(snippet) {
		t.Fatalf("MatchStart/MatchLen out of range: start=%d len=%d snippetBytes=%d",
			h.MatchStart, h.MatchLen, len(snippet))
	}
	gotMatch := snippet[h.MatchStart : h.MatchStart+h.MatchLen]
	if gotMatch != needle {
		t.Fatalf("match span = %q, want %q (start=%d len=%d)", gotMatch, needle, h.MatchStart, h.MatchLen)
	}

	assertContains(t, resp.Output, "  chat_history.jsonl:1:tool_result:")
	assertContains(t, resp.Output, needle)
	assertContains(t, resp.Output, "...")
	assertNotContains(t, resp.Output, "\x1b[")
}

func hitLens(resp *Response) []int {
	if len(resp.Matches) == 0 {
		return nil
	}
	out := make([]int, len(resp.Matches))
	for i, m := range resp.Matches {
		out[i] = len(m.Hits)
	}
	return out
}

func truncateForLog(s string, maxRunes int) string {
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	r := []rune(s)
	return string(r[:maxRunes]) + "…"
}
```
