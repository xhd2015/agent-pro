---
label: e2e
---

## Expected

- Paris then exited true.
- `resume.ready: yes` (or JSON ready true) after exit.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.HasParis {
		t.Fatalf("want Paris")
	}
	if !resp.ExitedTrue {
		t.Fatalf("want exited true:\n%s\njson:%s", resp.StatusAfterExit.Stdout, resp.StatusJSONAfterExit)
	}
	if !resp.ResumeReady {
		t.Fatalf("want resume.ready yes after exit:\n%s\njson:%s",
			resp.StatusAfterExit.Stdout, resp.StatusJSONAfterExit)
	}
}
```
