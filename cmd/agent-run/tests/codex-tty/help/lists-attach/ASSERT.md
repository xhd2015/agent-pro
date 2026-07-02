## Expected

- Exit code 0.
- Stdout usage lists `attach` subcommand.

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
	assertSuccess(t, resp)
	assert.Output(t, resp.Stdout, `` +
`<contains>
Usage:
attach
web
run
sessions
</contains>`)
}
```
