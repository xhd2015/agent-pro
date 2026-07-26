## Expected
- Same as notify-json-valid: ExitCode 0, event_id, partition, offset, received_at populated.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertSuccess(t, resp)
    obj := parseJSON(t, resp.Stdout)
    if _, ok := obj["event_id"].(string); !ok || obj["event_id"] == "" {
        t.Fatal("event_id is empty")
    }
    if n, ok := toInt(obj["offset"]); !ok || n < 0 {
        t.Fatal("offset missing or negative")
    }
}
```
