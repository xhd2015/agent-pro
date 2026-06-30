## Expected

- Exit code 1.
- Stderr indicates an unknown or unrecognized subcommand.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertExitCode(t, resp, 1)
}
```