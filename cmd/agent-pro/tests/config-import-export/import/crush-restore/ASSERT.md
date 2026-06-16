## Expected
- All crush files are restored to their correct disk locations.
- `crush/config/crush.json` → `~/.config/crush/crush.json`
- `crush/data/crush.json` → `~/.local/share/crush/crush.json`
- File contents match the original zip entries.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertFileExists(t, filepath.Join(req.HomeDir, ".config/crush/crush.json"))
	assertFileExists(t, filepath.Join(req.HomeDir, ".local/share/crush/crush.json"))
}
```
