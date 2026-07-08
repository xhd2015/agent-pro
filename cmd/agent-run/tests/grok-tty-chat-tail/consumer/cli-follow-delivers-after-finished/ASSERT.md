## Expected

- CLI stdout includes `CHAT_TAIL_CLI_FOLLOW_MARKER` from appended event.
- Demonstrates `sessions --print` follow path uses `WatchEvents` until ctx ends, not status gate.

## Exit Code

0 or timeout after appended line delivered (probe uses deadline)

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(resp.Stdout, chatTailCLIFollowMarker) {
		t.Fatalf("sessions --print did not tail appended event on finished session; stdout=%s stderr=%s",
			resp.Stdout, resp.Stderr)
	}
}
```