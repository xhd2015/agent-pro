---
label: e2e
---

## Expected

- Exit code 0.
- Run help lists `codex-tty` as a valid `--agent-runner` value.

## Exit Code

0

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	assert.Output(t, resp.Stdout, `` +
`<contains>
codex-tty
--agent-runner
</contains>`)
}
```
