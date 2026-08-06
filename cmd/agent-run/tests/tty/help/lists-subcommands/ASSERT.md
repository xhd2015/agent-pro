---
label: e2e
---

## Expected

- Exit code 0.
- Stdout lists `status`, `attach`, `send`, and `kill` subcommands.

## Expected Output

```
<contains>
<start-with>
  status
</start-with>
<start-with>
  attach
</start-with>
<start-with>
  send
</start-with>
<start-with>
  kill
</start-with>
</contains>
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertExitCode(t, resp, 0)
	assert.Output(t, resp.Stdout, ``+
		`<contains>
<start-with>
  status
</start-with>
<start-with>
  attach
</start-with>
<start-with>
  send
</start-with>
<start-with>
  kill
</start-with>
</contains>`)
}
```
