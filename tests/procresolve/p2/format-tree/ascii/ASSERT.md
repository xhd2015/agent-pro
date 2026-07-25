## Expected Output

```
200 /usr/local/bin/agent-run run --session-id=ignored-cli
+-- 201 /usr/local/bin/agent-run serve --session-id=ignored-cli
|   `-- 202 /usr/local/bin/grok
```

## Expected

- No error from Run.
- `TreeText` is a strict full match of the ASCII template above (trailing `\n`).
- Contains ASCII connectors: `+--`, `` `-- ``, and `|`.
- Does **not** contain Unicode box-drawing runes `├`, `└`, `│`, or `─`.

## Side Effects

- None.

## Errors

- None.

## Exit Code

N/A

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	if resp == nil {
		t.Fatal("response is nil")
	}
	// Avoid nested raw-string backticks: ASCII last-child connector is "`--".
	// v3 lines are raw Go regexps — escape + and |.
	lastConn := "`--" // ASCII tree last-child connector
	tmpl := "---\nversion: 3\n---\n" +
		"200 /usr/local/bin/agent-run run --session-id=ignored-cli\n" +
		`\+-- 201 /usr/local/bin/agent-run serve --session-id=ignored-cli` + "\n" +
		`\|   ` + lastConn + ` 202 /usr/local/bin/grok` + "\n"
	assert.Output(t, resp.TreeText, tmpl)
	for _, glyph := range []string{"├", "└", "│", "─"} {
		if strings.Contains(resp.TreeText, glyph) {
			t.Fatalf("ASCII TreeText must not contain Unicode connector rune %q:\n%s", glyph, resp.TreeText)
		}
	}
	if !strings.Contains(resp.TreeText, "+--") {
		t.Fatalf("TreeText missing +--:\n%s", resp.TreeText)
	}
	if !strings.Contains(resp.TreeText, lastConn) {
		t.Fatalf("TreeText missing %s:\n%s", lastConn, resp.TreeText)
	}
}
```
