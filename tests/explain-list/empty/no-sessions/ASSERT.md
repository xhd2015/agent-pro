## Expected Output

```
---
version: 3
---
No explain sessions yet\.
```

## Expected

- Exit code 0.
- Stdout is exactly `No explain sessions yet.\n` (trailing newline).
- No ANSI escapes.
- Fake agent not invoked (stderr has no `FAKE_AGENT_INVOKED`).

## Side Effects

- None (read-only list).

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
	assert.Output(t, resp.Stdout, `---
version: 3
---
No explain sessions yet\.
`)
}
```
