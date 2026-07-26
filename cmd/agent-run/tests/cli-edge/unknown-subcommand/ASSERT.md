## Expected

- Exit code 1 (L2: Handle returns unknown-command error).
- Stderr indicates an unknown or unrecognized subcommand.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertExitCode(t, resp, 1)
}
```