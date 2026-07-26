---
label: unit
explanation: --prefix matches channel name prefix only
---

## Expected Output

```
C0ALE44K5J6  #general  public  member
```

## Expected

- Exit code 0.
- Prefix match on `general` for QUERY `gen`.
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
C0ALE44K5J6  #general  public  member
`)
}
```
