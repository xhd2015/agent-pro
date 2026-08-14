## Expected

- `resp.InputBox` is `unknown` even though the unused fixture would be empty.
- `resp.HumanLine` is `input box: unknown`.
- `resp.JSONValue` is `unknown`.

## Exit Code

N/A (direct package call)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if !req.Unreachable {
		t.Fatal("this leaf must set Unreachable")
	}
	assertInputBox(t, resp, err, "unknown")
	if resp.JSONValue != "unknown" {
		t.Fatalf("JSONValue=%q want unknown (input_box)", resp.JSONValue)
	}
	assert.Output(t, resp.HumanLine+"\n", `---
version: 3
---
input box: unknown
`)
}
```
