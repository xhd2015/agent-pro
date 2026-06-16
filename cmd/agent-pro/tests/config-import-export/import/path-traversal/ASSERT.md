## Expected
- The path traversal entry is rejected (no file written outside home).
- Valid entries (without `..`) are still imported.
- No error for the valid entry.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertFileExists(t, filepath.Join(req.HomeDir, ".local/share/opencode/auth.json"))
	assertFileNotExist(t, filepath.Join(req.HomeDir, "../../etc/passwd"))
}
```
