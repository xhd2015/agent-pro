## Expected
- Behavior matches consolidated trace parsing semantics for `summary/title-from-identifier`.

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

	if resp.Output != "Shell Tool Call" { t.Fatalf("title = %q, want Shell Tool Call", resp.Output) }

}
```
