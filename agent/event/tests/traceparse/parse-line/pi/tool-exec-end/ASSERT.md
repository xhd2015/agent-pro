## Expected
- Behavior matches consolidated trace parsing semantics for `parse-line/pi/tool-exec-end`.

```go
import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}

	if !resp.OK { t.Fatal("expected parse ok") }
	assertContains(t, resp.Output, `"subtype":"completed"`)
	assertContains(t, resp.Output, `"status":"completed"`)
	assertContains(t, resp.Output, "done")

}
```
