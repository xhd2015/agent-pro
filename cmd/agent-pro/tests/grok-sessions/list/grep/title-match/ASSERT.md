## Expected Output

```
---
version: 3
---
SESSION ID                              LAST ACTIVE   TITLE                                        MSGS  CWD
01900010-aaaa-7aaa-aaaa-aaaaaaaaaaaa    30m ago       Implement GREP_TITLE_TOKEN feature              0  /tmp/grep-title-match
  summary\.json:1:title: Implement GREP_TITLE_TOKEN feature
```

## Expected

- Exactly one matching session: `01900010-aaaa-7aaa-aaaa-aaaaaaaaaaaa`.
- Non-matching session `01900010-bbbb-...` is omitted.
- Output includes indented hit `summary.json:1:title:` containing `GREP_TITLE_TOKEN`.
- No ANSI escapes (color never).
- Package formatter output matches classic table spacing (no forced trailing `\n` after last line; same TrimRight style as FormatListTable).

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if len(resp.Sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(resp.Sessions))
	}
	if resp.Sessions[0].ID != "01900010-aaaa-7aaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("session id = %q, want 01900010-aaaa-7aaa-aaaa-aaaaaaaaaaaa", resp.Sessions[0].ID)
	}
	assertNotContains(t, resp.Output, "01900010-bbbb-7bbb-bbbb-bbbbbbbbbbbb")
	assertNotContains(t, resp.Output, "\x1b[")

	// No trailing newline: matches FormatListTable TrimRight style.
	assert.Output(t, resp.Output, `---
version: 3
---
SESSION ID                              LAST ACTIVE   TITLE                                        MSGS  CWD
01900010-aaaa-7aaa-aaaa-aaaaaaaaaaaa    30m ago       Implement GREP_TITLE_TOKEN feature              0  /tmp/grep-title-match
  summary\.json:1:title: Implement GREP_TITLE_TOKEN feature`)
}
```
