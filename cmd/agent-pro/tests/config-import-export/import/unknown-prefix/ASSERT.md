## Expected
- Entries with unknown prefixes (`cursor/`) are silently skipped.
- Known prefix entries (`opencode/`) are imported normally.
- No error.

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	assertFileExists(t, filepath.Join(req.HomeDir, ".local/share/opencode/auth.json"))
	assertFileNotExist(t, filepath.Join(req.HomeDir, ".cursor/settings.json"))
}
```
