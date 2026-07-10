## Expected

- Exit code 0.
- Stdout matches same usage as `channels -h`.
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
version: 2
---
slack-msg channels: list or search workspace channels.

Usage:
  slack-msg channels <command> [options]

Commands:
  list    List visible channels
  search  Search channels by name

Options:
  -h, --help  Show help
`)
}
```
