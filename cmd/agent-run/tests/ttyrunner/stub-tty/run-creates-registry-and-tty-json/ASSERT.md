## Expected

- Registry and tty.json created after stub-tty run.

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertFileExists(t, resp.RegistryPath)
	assertFileExists(t, resp.TTYJSONPath)
}
```
