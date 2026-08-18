---
label: e2e
---

## Expected Output

```
---
version: 3
---
Recent explain sessions \(1 shown of 1, limit 10\)

── 1 ──  2026-07-13 09:00:00  ·  opencode / deepseek-chat  ·  1 turn
   Q  valid question
   A  valid answer
```

## Expected

- Exit 0.
- Only the valid session is listed (`1 shown`).
- Corrupt dirname / content not required to appear; no hard failure.
- Trailing newline; no ANSI.

## Side Effects

- None.

## Errors

- None (silent skip).

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
	assertContains(t, resp.Stdout, "valid question")
	assertNotContains(t, resp.Stdout, "corrupt")
	assert.Output(t, resp.Stdout, `---
version: 3
---
Recent explain sessions \(1 shown of 1, limit 10\)

── 1 ──  2026-07-13 09:00:00  ·  opencode / deepseek-chat  ·  1 turn
   Q  valid question
   A  valid answer
`)
}
```
