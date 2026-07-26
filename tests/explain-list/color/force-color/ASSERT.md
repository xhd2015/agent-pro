---
label: e2e
---

## Expected Output

Colored card (labels + dim header; bodies plain):

```
---
version: 2
---
Recent explain sessions (1 shown of 1, limit 10)

<ansi-color #2>── 1 ──  2026-07-13 14:30:05  ·  opencode / deepseek-chat  ·  1 turn</ansi-color>
   <ansi-color #1;36>Q</ansi-color>  color q
   <ansi-color #1;32>A</ansi-color>  color a
```

## Expected

- Exit 0.
- Stdout contains ANSI CSI sequences.
- Q label uses bold cyan `\x1b[1;36m`; A uses bold green `\x1b[1;32m`; header dim `\x1b[2m`.
- Bodies `color q` / `color a` appear as plain text (not wrapped alone in color in the template).
- Trailing newline.

## Side Effects

- None.

## Errors

- None.

## Exit Code

- 0.

```go
import (
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
	assertHasANSI(t, resp.Stdout)

	// Locked SGR codes from requirement.
	assertContains(t, resp.Stdout, "\x1b[1;36m") // bold cyan Q
	assertContains(t, resp.Stdout, "\x1b[1;32m") // bold green A
	assertContains(t, resp.Stdout, "\x1b[2m")    // dim meta
	assertContains(t, resp.Stdout, "\x1b[0m")    // reset

	// Bodies present as plain text next to colored labels.
	assertContains(t, resp.Stdout, "color q")
	assertContains(t, resp.Stdout, "color a")

	// Full structured color template (labels only colored; bodies plain).
	assert.Output(t, resp.Stdout, `---
version: 2
---
Recent explain sessions (1 shown of 1, limit 10)

<ansi-color #2>── 1 ──  2026-07-13 14:30:05  ·  opencode / deepseek-chat  ·  1 turn</ansi-color>
   <ansi-color #1;36>Q</ansi-color>  color q
   <ansi-color #1;32>A</ansi-color>  color a
`)
}
```
