---
label: unit
explanation: multi-word MESSAGE positional
---

## Expected

- Exit code 0.
- Sending line quotes full multi-word message.

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
version: 3
__TS__: type=number, example=1783398010.628649, message timestamp
---
Sending to channel=C0ALE44K5J6: "Hello from script"
Using config from: \(none\)
OK ts=__TS__ channel=C0ALE44K5J6
`)
}
```
