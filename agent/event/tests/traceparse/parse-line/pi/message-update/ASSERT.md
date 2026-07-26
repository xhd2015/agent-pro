## Expected
- Behavior matches consolidated trace parsing semantics for `parse-line/pi/message-update`.

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
	assertContains(t, resp.Output, `"role":"assistant"`)
	assertContains(t, resp.Output, "hello delta")
	assertNotContains(t, resp.Output, "world")

}
```
