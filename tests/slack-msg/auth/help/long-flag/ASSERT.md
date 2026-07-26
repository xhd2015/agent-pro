## Expected

- Exit code 0.
- Stdout matches same `auth` usage as `-h` (lists `status`).
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
slack-msg auth: inspect bot or app token status.

Usage:
  slack-msg auth <command> [options]

Commands:
  status  Show bot or app token status

Options:
  -h, --help  Show help
`)
}
```
