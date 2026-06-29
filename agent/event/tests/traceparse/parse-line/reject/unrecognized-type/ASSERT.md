## Expected
- Behavior matches consolidated trace parsing semantics for `parse-line/reject/unrecognized-type`.

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

	if resp.OK { t.Fatal("expected no adapter match") }

}
```
