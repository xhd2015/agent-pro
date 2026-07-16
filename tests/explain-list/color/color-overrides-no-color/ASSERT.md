## Expected

- Exit 0.
- ANSI present despite `NO_COLOR=1`.
- Contains bold cyan Q (`\x1b[1;36m`), bold green A (`\x1b[1;32m`), dim (`\x1b[2m`).
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
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 0)
	assertStdoutEndsWithNewline(t, resp.Stdout)
	assertHasANSI(t, resp.Stdout)
	assertContains(t, resp.Stdout, "\x1b[1;36m")
	assertContains(t, resp.Stdout, "\x1b[1;32m")
	assertContains(t, resp.Stdout, "\x1b[2m")
	assert.Output(t, resp.Stdout, `---
version: 3
---
Recent explain sessions \(1 shown of 1, limit 10\)

<ansi-color #2>── 1 ──  2026-07-13 14:30:05  ·  opencode / deepseek-chat  ·  1 turn</ansi-color>
   <ansi-color #1;36>Q</ansi-color>  color q
   <ansi-color #1;32>A</ansi-color>  color a
`)
}
```
