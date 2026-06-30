## Expected

- `Attach` returns non-nil error.
- Error message mentions interactive terminal or TTY requirement.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.AttachErr == "" {
		t.Fatal("expected attach error for non-TTY stdin")
	}
	lower := strings.ToLower(resp.AttachErr)
	assert.Output(t, lower, `
<contains>
interactive
terminal
</contains>`)
}
```