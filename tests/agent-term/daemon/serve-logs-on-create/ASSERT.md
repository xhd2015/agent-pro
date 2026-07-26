## Expected

- After `POST /api/terminal/sessions`, daemon stderr logs the HTTP method and path.
- Session id from create response is available to the harness.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(resp.HTTPBody, "session-") {
		t.Fatalf("expected session id from create, got %q", resp.HTTPBody)
	}
	assert.Output(t, resp.DaemonStderr, `
<contains>
POST
/api/terminal/sessions
</contains>`)
}
```