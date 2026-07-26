## Expected
- Roundtripped output preserves ActionMessage type.
- Text "Hello world" appears in the output (from msg_start and msg_update events).
- After fix, FromPi uses Delta for message_update events; for instant phase, ToPi sets Delta = Text,
  so the roundtrip correctly preserves the text through the delta path.
- msg_end event should have empty Text (deltas already shown via message_update).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertContains(t, resp.Output, `"type":"message"`)
	assertContains(t, resp.Output, `"Hello world"`)
}
```
