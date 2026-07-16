---
label: unit
explanation: --thread replies path prints chronological human lines
---

## Expected Output

```
[1710001000.000100] U_PARENT: parent
[1710001000.000200] U_R1: reply one
[1710001000.000300] U_R2: reply two
```

## Expected

- Exit code 0.
- Three chronological lines from replies mock.
- Stderr empty.

## Exit Code

0

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
	if resp.Stderr != "" {
		t.Fatalf("expected empty stderr, got:\n%s", resp.Stderr)
	}
	assert.Output(t, resp.Stdout, `---
version: 3
---
\[1710001000\.000100\] U_PARENT: parent
\[1710001000\.000200\] U_R1: reply one
\[1710001000\.000300\] U_R2: reply two
`)
}
```
