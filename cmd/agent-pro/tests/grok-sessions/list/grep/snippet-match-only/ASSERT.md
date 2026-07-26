## Expected

- One matching session: `01900022-aaaa-7aaa-aaaa-aaaaaaaaaaaa`.
- One hit on `chat_history.jsonl:1:assistant:`.
- Pattern and field match are `1100×'X'`; after windowing:
  - `Hits[0].Snippet` has **exactly 1024 runes** (first 1024 of the match).
  - Snippet equals `1024×'X'` (no side context outside the match).
  - Snippet does **not** contain the full untruncated `1100×'X'` body.
- `MatchStart == 0` and `MatchLen == len(Snippet)` (entire snippet is the truncated match).
- Rune counts use `utf8.RuneCountInString`.
- No ANSI escapes.

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
	if resp.Sessions[0].ID != "01900022-aaaa-7aaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("session id = %q, want 01900022-aaaa-7aaa-aaaa-aaaaaaaaaaaa", resp.Sessions[0].ID)
	}
	if len(resp.Matches) != 1 || len(resp.Matches[0].Hits) != 1 {
		t.Fatalf("want 1 match with 1 hit, got matches=%d", len(resp.Matches))
	}

	h := resp.Matches[0].Hits[0]
	if h.File != "chat_history.jsonl" || h.Line != 1 || h.Part != "assistant" {
		t.Fatalf("hit meta = %s:%d:%s, want chat_history.jsonl:1:assistant", h.File, h.Line, h.Part)
	}

	const wantRunes = 1024
	snippet := h.Snippet
	rc := utf8.RuneCountInString(snippet)
	if rc > wantRunes {
		t.Fatalf("snippet rune count = %d, want ≤ %d (match-only cap)", rc, wantRunes)
	}
	if rc != wantRunes {
		t.Fatalf("snippet rune count = %d, want exactly %d (first 1024 of match)", rc, wantRunes)
	}
	wantSnippet := strings.Repeat("X", wantRunes)
	if snippet != wantSnippet {
		t.Fatalf("snippet = %q (len runes=%d), want %d×'X'", truncateMatchLog(snippet), rc, wantRunes)
	}
	// Full field would be 1100 X's; must not dump past the cap.
	if strings.Contains(snippet, strings.Repeat("X", 1100)) || rc >= 1100 {
		t.Fatalf("snippet still holds full match body (≥1100), not truncated")
	}
	if h.MatchStart != 0 {
		t.Fatalf("MatchStart = %d, want 0 (snippet is truncated match only)", h.MatchStart)
	}
	if h.MatchLen != len(snippet) {
		t.Fatalf("MatchLen = %d, want %d (entire snippet is the match span)", h.MatchLen, len(snippet))
	}

	assertContains(t, resp.Output, "  chat_history.jsonl:1:assistant:")
	assertNotContains(t, resp.Output, "\x1b[")
}

func truncateMatchLog(s string) string {
	if utf8.RuneCountInString(s) <= 40 {
		return s
	}
	return string([]rune(s)[:20]) + "…" + string([]rune(s)[utf8.RuneCountInString(s)-5:])
}
```
