---
label: unit
explanation: multi-page conversations.list merged, archived excluded, name-sorted
---

## Expected Output

```
C0AGENTDBG1  #agent-pro-debug  private  -
C0ALE44K5J6  #general  public  member
C0OTHERCHAN  #random  public  -
```

## Expected

- Exit code 0.
- Same sorted non-archived human lines as single-page list (pages merged).
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
C0OTHERCHAN  #random  public  -
`)
}
```
