## Expected

- Command exits non-zero promptly (no hang waiting for interactive bash).
- Error message mentions interactive terminal or TTY requirement (same as attach).

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for non-TTY run, combined %q", resp.Combined)
	}
	lower := strings.ToLower(resp.Combined)
	assert.Output(t, lower, `
<contains>
interactive
terminal
</contains>`)
}
```