---
label: e2e
---

## Expected Output

```
---
version: 3
---
No matching explain sessions\.
```

## Expected

- Exit 0.
- Stdout exactly `No matching explain sessions.\n`.
- Not the empty-store message; no session markers; no ANSI.
- Fake agent not invoked.

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
	assertNotContains(t, resp.Stderr, "FAKE_AGENT_INVOKED")
	assertNotContains(t, resp.Stdout, "No explain sessions yet")
	assertNotContains(t, resp.Stdout, "marker-cat")
	assert.Output(t, resp.Stdout, `---
version: 3
---
No matching explain sessions\.
`)
}
```
