---
label: unit
explanation: direct group ID passthrough
---

## Expected

- Exit code 0.
- Sending line uses `G024BE91L` unchanged.

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
__TS__: type=number, example=1783398010.628649, message timestamp
---
Sending to channel=G024BE91L: "direct G"
Using config from: (none)
OK ts=__TS__ channel=G024BE91L
`)
}
```