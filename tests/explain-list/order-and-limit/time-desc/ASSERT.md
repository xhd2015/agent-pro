---
label: e2e
---

## Expected Output

```
---
version: 2
---
Recent explain sessions (2 shown of 2, limit 10)

── 1 ──  2026-07-13 14:30:05  ·  opencode / deepseek-chat  ·  1 turn
   Q  newer question
   A  newer answer

── 2 ──  2026-07-12 09:15:22  ·  opencode / deepseek-chat  ·  1 turn
   Q  older question
   A  older answer
```

## Expected

- Exit 0.
- Newer session is index 1; older is index 2.
- Title reports 2 shown, limit 10.
- No ANSI (plain list).
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
	assertNoANSI(t, resp.Stdout)
	assert.Output(t, resp.Stdout, `---
version: 2
---
Recent explain sessions (2 shown of 2, limit 10)

── 1 ──  2026-07-13 14:30:05  ·  opencode / deepseek-chat  ·  1 turn
   Q  newer question
   A  newer answer

── 2 ──  2026-07-12 09:15:22  ·  opencode / deepseek-chat  ·  1 turn
   Q  older question
   A  older answer
`)
}
```
