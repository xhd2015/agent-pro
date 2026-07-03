## Expected

- Run duration reflects banner delay (>= 600ms).

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.ExitCode != 0 { t.Fatalf("exit %d: %s", resp.ExitCode, resp.Stderr) }
}
```
