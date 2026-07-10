---
label: unit
explanation: default contains match; case-insensitive; strip leading # from query
---

## Expected Output

```
C0ALE44K5J6  #general  public  member
```

## Expected

- Exit code 0.
- Single matching line for `#general` (QUERY `#GENERAL` contains-matches `general` after # strip + case fold).
- Does not match unrelated channels (including `agent-pro-debug`, which only contains substring `gen` not `general`); stderr empty; trailing newline.

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
version: 2
---
C0ALE44K5J6  #general  public  member
`)
}
```
