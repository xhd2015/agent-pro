## Expected

- Kind is tool_calling and detail is bash.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.Kind != "tool_calling" || resp.Detail != "bash" {
		t.Fatalf("%+v", resp)
	}
}
```
