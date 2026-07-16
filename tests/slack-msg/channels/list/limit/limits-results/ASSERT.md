---
label: unit
explanation: --limit 2 prints first two channels by name after filter/sort
---

## Expected Output

```
C0AGENTDBG1  #agent-pro-debug  private  -
C0ALE44K5J6  #general  public  member
```

## Expected

- Exit code 0.
- Exactly two human lines (first two of sorted non-archived set).
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
C0AGENTDBG1  #agent-pro-debug  private  -
C0ALE44K5J6  #general  public  member
`)
}
```
