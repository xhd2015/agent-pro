## Expected
- Behavior matches consolidated trace parsing semantics for `summary/title-from-identifier`.

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

	if resp.Output != "Shell Tool Call" { t.Fatalf("title = %q, want Shell Tool Call", resp.Output) }

}
```
