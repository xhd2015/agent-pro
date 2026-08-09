## Expected

- Set contains `PATH=…`.
- PATH value starts with `/opt/agent-bin-a` + path list separator + `/opt/agent-bin-b`.
- Unset is empty.

## Errors

- Missing PATH; prefixes reversed or not joined first.

```go
import (
	"os"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertBuildOK(t, resp, err)
	assertUnsetEmpty(t, resp.Unset)
	pathVal, ok := setGet(resp.Set, "PATH")
	if !ok {
		t.Fatalf("Set missing PATH; Set=%#v", resp.Set)
	}
	sep := string(os.PathListSeparator)
	wantPrefix := "/opt/agent-bin-a" + sep + "/opt/agent-bin-b"
	if pathVal != wantPrefix && !strings.HasPrefix(pathVal, wantPrefix+sep) {
		t.Fatalf("PATH=%q, want prefix %q (optionally + sep + parent PATH)", pathVal, wantPrefix)
	}
}
```
