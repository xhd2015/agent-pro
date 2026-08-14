## Expected

- `resp.InputBox` is `empty`.
- `resp.HumanLine` is exactly `input box: empty` (CLI printers append the
  trailing `\n` so the last content line is newline-terminated).
- `resp.JSONValue` is `empty` (JSON field `input_box`).

## Exit Code

N/A (direct package call)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertInputBox(t, resp, err, "empty")
	if resp.JSONValue != "empty" {
		t.Fatalf("JSONValue=%q want empty (input_box)", resp.JSONValue)
	}
	assert.Output(t, resp.HumanLine+"\n", `---
version: 3
---
input box: empty
`)
}
```
