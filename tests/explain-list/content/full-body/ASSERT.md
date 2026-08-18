---
label: e2e
---

## Expected Output

```
---
version: 3
---
Recent explain sessions \(1 shown of 1, limit 10\)

── 1 ──  2026-07-13 10:00:00  ·  opencode / deepseek-chat  ·  1 turn
   Q  short q
   A  <200 ASCII x characters, full, no ellipsis>
```

## Expected

- Exit 0.
- Stdout contains the full untruncated answer (`strings.Repeat("x", 200)`).
- Stdout does **not** contain ellipsis `…`.
- Exact card template: `Q  short q` then `A  ` + 200 `x`.
- Trailing newline; no ANSI.

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

	full := strings.Repeat("x", 200)
	assertContains(t, resp.Stdout, full)
	assertNotContains(t, resp.Stdout, "…")

	// Build template without embedding 200 x in a raw string (readability).
	// Trailing "" in Join yields the final \n required for CLI stdout.
	template := strings.Join([]string{
		"---",
		"version: 3",
		"---",
		"Recent explain sessions \\(1 shown of 1, limit 10\\)",
		"",
		"── 1 ──  2026-07-13 10:00:00  ·  opencode / deepseek-chat  ·  1 turn",
		"   Q  short q",
		"   A  " + full,
		"",
	}, "\n")
	assert.Output(t, resp.Stdout, template)
}

```
