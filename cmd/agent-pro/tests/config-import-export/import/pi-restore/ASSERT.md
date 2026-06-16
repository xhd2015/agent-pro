## Expected
- All pi files are restored to their correct disk locations under `~/.pi/agent/`.
- File contents match the original zip entries.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertFileExists(t, filepath.Join(req.HomeDir, ".pi/agent/auth.json"))
	assertFileExists(t, filepath.Join(req.HomeDir, ".pi/agent/settings.json"))
	assertFileExists(t, filepath.Join(req.HomeDir, ".pi/agent/models.json"))
}
```
