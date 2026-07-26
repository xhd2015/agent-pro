---
label: e2e
---

## Expected Output

```
---
version: 2
---
Recent explain sessions (1 shown of 1, limit 10)

── 1 ──  2026-07-13 14:30:05  ·  opencode / deepseek-chat  ·  2 turns
   Q  What is a goroutine?
   A  A goroutine is a lightweight thread managed by the Go runtime.
   Q  How does the scheduler work?
   A  The Go scheduler multiplexes goroutines onto OS threads.
```

## Expected

- Exit 0.
- Title: 1 shown, limit 10.
- Header: formatted time, runner/model, **2 turns** (user message count).
- Four Q/A lines in message order; labels `Q` / `A` with two spaces before body.
- Trailing newline; no ANSI.

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
Recent explain sessions (1 shown of 1, limit 10)

── 1 ──  2026-07-13 14:30:05  ·  opencode / deepseek-chat  ·  2 turns
   Q  What is a goroutine?
   A  A goroutine is a lightweight thread managed by the Go runtime.
   Q  How does the scheduler work?
   A  The Go scheduler multiplexes goroutines onto OS threads.
`)
}
```
