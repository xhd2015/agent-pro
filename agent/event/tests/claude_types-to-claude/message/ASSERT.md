## Expected
- Output JSON array contains one `assistant` event with a `text` content block whose text is `pong`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"assistant"`)
	assertContains(t, resp.Output, `"type":"text"`)
	assertContains(t, resp.Output, `"text":"pong"`)
}
```
