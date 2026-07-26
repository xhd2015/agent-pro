## Expected

- Command exits non-zero.
- Stderr mentions `agent-term serve`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatal("expected non-zero exit when daemon down")
	}
	assert.Output(t, resp.Combined, `
<contains>
agent-term serve
</contains>`)
}
```