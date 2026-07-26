---
label: e2e
---

## Expected

- Attach stdout contains `ATTACH_STDIN_MARKER` after typing through attach.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.Stdout, "ATTACH_STDIN_MARKER") {
		t.Fatalf("tty-watch attach missing stdin marker; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
}
```
