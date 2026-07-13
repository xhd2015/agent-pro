## Expected

- Exit 0.
- Stdout lists the seeded session (`no llm q` / `no llm a`).
- Stderr does not contain `FAKE_AGENT_INVOKED`.
- Exit code is not 99 (fake agent exit).
- Trailing newline; no ANSI.

## Side Effects

- None (read-only).

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
	assertNotContains(t, resp.Stderr, "FAKE_AGENT_INVOKED")
	assertNotContains(t, resp.Stdout, "FAKE_AGENT_INVOKED")
	assertContains(t, resp.Stdout, "no llm q")
	assertContains(t, resp.Stdout, "no llm a")
	assert.Output(t, resp.Stdout, `---
version: 2
---
Recent explain sessions (1 shown of 1, limit 10)

── 1 ──  2026-07-13 12:00:00  ·  opencode / deepseek-chat  ·  1 turn
   Q  no llm q
   A  no llm a
`)
}
```
