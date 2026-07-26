---
label: unit
explanation: "search soft-skip private; public contains hit + warning with see: topic exit 0"
---

## Expected Output

```
C0ALE44K5J6  #general  public  member
```

## Expected

- Exit code 0.
- Single human line for `#general`.
- Stderr contains
  `warning: skipped private channels (missing groups:read); see: slack-msg --help --topic add-missing-scope`.
- Trailing newline.

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
	assertStderrContains(t, resp, "warning: skipped private channels (missing groups:read); see: slack-msg --help --topic add-missing-scope")
	assert.Output(t, resp.Stdout, `---
version: 2
---
C0ALE44K5J6  #general  public  member
`)
}
```
