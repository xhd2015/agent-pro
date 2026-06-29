## Expected
- Behavior matches consolidated trace parsing semantics for `thin-wrapper/parse-reexport`.

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}

	assertContains(t, resp.Output, `"tool_name":"Config Warning"`)

}
```
