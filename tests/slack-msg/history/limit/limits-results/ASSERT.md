---
label: unit
explanation: --limit 2 returns two newest messages printed oldest→newest
---

## Expected Output

```
[1710000002.000200] U_NEWER: second message
[1710000003.000300] U_NEWEST: third message
```

## Expected

- Exit code 0.
- Exactly two human lines (the two newest, chronological among them).
- Stderr empty.

## Exit Code

0

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
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	assert.Output(t, resp.Stdout, `---
version: 2
---
[1710000002.000200] U_NEWER: second message
[1710000003.000300] U_NEWEST: third message
`)
}
```
