## Expected
- Output JSON array contains one `assistant` event with a `thinking` content block whose `thinking` text is `reasoning`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertContains(t, resp.Output, `"type":"assistant"`)
	assertContains(t, resp.Output, `"type":"thinking"`)
	assertContains(t, resp.Output, `"thinking":"reasoning"`)
}
```
