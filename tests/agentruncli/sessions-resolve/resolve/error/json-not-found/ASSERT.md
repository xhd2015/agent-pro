## Errors

- Not-found error mentioning the UUID.
- Stdout must not be a successful JSON object (empty preferred).

```go
import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoRunError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	uuid := "99999999-9999-9999-9999-999999999999"
	want := fmt.Sprintf("session not found: no grok session with runner_session_id %s", uuid)
	assertExactErr(t, resp.Err, want)
	trim := strings.TrimSpace(resp.Stdout)
	if trim == "" {
		return
	}
	var obj map[string]any
	if json.Unmarshal([]byte(trim), &obj) == nil {
		if _, ok := obj["session_id"]; ok {
			t.Fatalf("--json not-found must not emit success JSON body, got %s", resp.Stdout)
		}
	}
}
```
