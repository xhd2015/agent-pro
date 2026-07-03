## Expected

- stub-tty run completes after screen frame sequence.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.ExitCode != 0 { t.Fatalf("exit %d", resp.ExitCode) }
}
```
