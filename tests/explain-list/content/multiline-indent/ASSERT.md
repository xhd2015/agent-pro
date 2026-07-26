---
label: e2e
---

## Expected Output

```
---
version: 2
---
Recent explain sessions (1 shown of 1, limit 10)

── 1 ──  2026-07-13 11:00:00  ·  opencode / deepseek-chat  ·  1 turn
   Q  hello

      world
   A  first
      second

      third
```

## Expected

- Exit 0.
- Newlines preserved (not collapsed to spaces).
- Continuation non-empty lines start with exactly **6 spaces** (align under body after `Q  ` / `A  `).
- Blank message segments emit pure `\n` (no spaces on the blank line).
- Exact template match; trailing newline; no ANSI.

## Side Effects

- None.

## Errors

- None.

## Exit Code

- 0.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	assertStdoutEndsWithNewline(t, resp.Stdout)
	assertNoANSI(t, resp.Stdout)

	// Must not collapse multi-line Q into a single "hello world" line.
	assertNotContains(t, resp.Stdout, "Q  hello world")

	// Continuation lines: exactly 6 spaces before non-empty body segments.
	assertContains(t, resp.Stdout, "      world")
	assertContains(t, resp.Stdout, "      second")
	assertContains(t, resp.Stdout, "      third")

	// Spot-check pure blank lines between indented segments (no spaces-only blank).
	if strings.Contains(resp.Stdout, "\n      \n") {
		t.Fatalf("blank continuation must be pure \\n, not spaces-only line:\n%s", resp.Stdout)
	}

	assert.Output(t, resp.Stdout, `---
version: 2
---
Recent explain sessions (1 shown of 1, limit 10)

── 1 ──  2026-07-13 11:00:00  ·  opencode / deepseek-chat  ·  1 turn
   Q  hello

      world
   A  first
      second

      third
`)
}
```
