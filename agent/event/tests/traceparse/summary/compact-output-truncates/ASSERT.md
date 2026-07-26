## Expected
- Behavior matches consolidated trace parsing semantics for `summary/compact-output-truncates`.

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

	assertContains(t, resp.Output, "[omitted:")
	assertContains(t, resp.Output, "--- first lines ---")
	assertContains(t, resp.Output, "--- last lines ---")

}
```
