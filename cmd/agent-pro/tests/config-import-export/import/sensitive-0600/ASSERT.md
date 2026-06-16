## Expected
- Files that may contain secrets get 0600 permission:
  - `pi/auth.json` → 0600
  - `opencode/opencode.jsonc` → 0600
  - `crush/config/crush.json` → 0600
- Non-sensitive files still get 0644.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertFileMode(t, filepath.Join(req.HomeDir, ".pi/agent/auth.json"), 0600)
	assertFileMode(t, filepath.Join(req.HomeDir, ".config/opencode/opencode.jsonc"), 0600)
	assertFileMode(t, filepath.Join(req.HomeDir, ".config/crush/crush.json"), 0600)
	assertFileMode(t, filepath.Join(req.HomeDir, ".local/share/opencode/settings.json"), 0644)
}
```
