## Expected

- `ResultArgs` deep-equal to `BaseArgs` (same length and tokens).
- No `--event-bus-url` or `--event-bus-token` present.

## Errors

- None.

```go
import (
	"reflect"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !reflect.DeepEqual(resp.ResultArgs, req.BaseArgs) {
		t.Fatalf("empty URL must leave args unchanged\n got %#v\nwant %#v", resp.ResultArgs, req.BaseArgs)
	}
	for _, a := range resp.ResultArgs {
		if a == "--event-bus-url" || a == "--event-bus-token" {
			t.Fatalf("unexpected event-bus flag in result: %#v", resp.ResultArgs)
		}
	}
}
```
