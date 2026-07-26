## Expected
- ExitCode != 0, error about --runner.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if resp.ExitCode == 0 {
        t.Fatal("expected non-zero exit code")
    }
    assertContains(t, resp.Stderr, "--runner")
}
```
