---
label: unit
explanation: session history --after-msg-id returns only later log lines
---

## Expected Output

```
[1710000902.000200] U2: second
[1710000903.000300] U1: third
```

## Expected

- Exit code 0.
- Only messages after `m1` (second + third).
- Stderr empty; trailing `\n`.

## Exit Code

0

```go
import (
	"strings"
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
	if strings.Contains(resp.Stdout, "first") {
		t.Fatalf("after-msg-id m1 must not include first message:\n%s", resp.Stdout)
	}
	assert.Output(t, resp.Stdout, `---
version: 3
---
[1710000902.000200] U2: second
[1710000903.000300] U1: third
`)
}
```
