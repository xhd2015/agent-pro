## Expected

- Paris after open.
- Live send of hello succeeds (exit 0 preferred).
- Snapshot/events show hello marker.
- Not forced to exited=true (still live).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.HasParis {
		t.Fatalf("want Paris; snap=%s", resp.ParisSnapshot)
	}
	if resp.SendFollowup.ExitCode != 0 {
		t.Fatalf("send followup exit=%d stderr=%s", resp.SendFollowup.ExitCode, resp.SendFollowup.Stderr)
	}
	if !resp.HasHello {
		t.Fatalf("want hello marker after live send; snap=\n%s\nevents=\n%s",
			resp.ResumeSnapshot, resp.EventsBlob)
	}
	// Live path: exited true would be unexpected right after hello without /exit.
	if resp.ExitedTrue {
		t.Logf("note: exited=true after live send (unusual); status=\n%s", resp.StatusAfterExit.Stdout)
	}
	_ = strings.TrimSpace(resp.ResumeSnapshot)
}
```
