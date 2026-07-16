## Expected Output

```
---
version: 3
---
Recent explain sessions \(1 shown of 1, limit 10\)

── 1 ──  2026-07-13 14:30:05  ·  opencode / deepseek-chat  ·  1 turn
   Q  color q
   A  color a
```

## Expected

- Exit 0.
- No ANSI in stdout (non-TTY auto off).
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
	assertNoANSI(t, resp.Stdout)
	assert.Output(t, resp.Stdout, `---
version: 3
---
Recent explain sessions \(1 shown of 1, limit 10\)

── 1 ──  2026-07-13 14:30:05  ·  opencode / deepseek-chat  ·  1 turn
   Q  color q
   A  color a
`)
}
```
