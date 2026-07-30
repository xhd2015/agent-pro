## Expected Output

```
---
version: 3
---
SESSION ID                              LAST ACTIVE   TITLE                                        MSGS  CWD
01900019-aaaa-7aaa-aaaa-aaaaaaaaaaaa    30m ago       Enable GREP_COLOR_TOKEN highlighting            0  /tmp/grep-color-never
  summary\.json:1:title: Enable GREP_COLOR_TOKEN highlighting
```

## Expected

- Session is listed with plain (non-ANSI) hit line.
- Output contains zero CSI escape sequences (`\x1b[`).

## Errors

- None.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	if len(resp.Sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(resp.Sessions))
	}
	if strings.Contains(resp.Output, "\x1b[") {
		t.Fatalf("unexpected ANSI escapes with Color=never:\n%s", resp.Output)
	}
	assertContains(t, resp.Output, "  summary.json:1:title: Enable GREP_COLOR_TOKEN highlighting")

	// No trailing newline: matches FormatListTable TrimRight style.
	assert.Output(t, resp.Output, `---
version: 3
---
SESSION ID                              LAST ACTIVE   TITLE                                        MSGS  CWD
01900019-aaaa-7aaa-aaaa-aaaaaaaaaaaa    30m ago       Enable GREP_COLOR_TOKEN highlighting            0  /tmp/grep-color-never
  summary\.json:1:title: Enable GREP_COLOR_TOKEN highlighting`)
}
```
