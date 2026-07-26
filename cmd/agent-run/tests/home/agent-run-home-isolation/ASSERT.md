---
label: e2e
---

## Expected

- Exit code 0.
- Every file created under `req.TempDir` during the run lives inside `AGENT_RUN_HOME`
  (excluding the built binaries under `bin/`).

```go
import (
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertSuccess(t, resp)
	binDir := filepath.Join(req.TempDir, "bin")
	outside := filesOutsidePrefix(t, req.TempDir, req.Home)
	var unexpected []string
	for _, p := range outside {
		if strings.HasPrefix(p, binDir+string(filepath.Separator)) || p == binDir {
			continue
		}
		unexpected = append(unexpected, p)
	}
	if len(unexpected) > 0 {
		t.Fatalf("files written outside AGENT_RUN_HOME:\n%s", strings.Join(unexpected, "\n"))
	}
}
```