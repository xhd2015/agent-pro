---
label: e2e
---

## Expected

- Exit code 0.
- Fake grok output contains `GROK_ARGV=-p hello` (argv after `grok` unchanged).
- `--log-http` path did not consume `-p` or `hello`.

## Exit Code

0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "GROK_ARGV=-p hello")
}
```