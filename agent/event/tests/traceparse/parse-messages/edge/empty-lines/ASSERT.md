## Expected
- Behavior matches consolidated trace parsing semantics for `parse-messages/edge/empty-lines`.

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

	if resp.Output != "[]" && resp.Output != "null" {
		var msgs []any
		_ = json.Unmarshal([]byte(resp.Output), &msgs)
		if len(msgs) != 0 { t.Fatalf("want empty messages, got %s", resp.Output) }
	}

}
```
