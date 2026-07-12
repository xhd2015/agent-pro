## Expected

- Paris after open.
- Both live sends exit 0.
- Hello marker present after followups.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.HasParis {
		t.Fatalf("want Paris")
	}
	if resp.SendFollowup.ExitCode != 0 {
		t.Fatalf("first send exit=%d: %s", resp.SendFollowup.ExitCode, resp.SendFollowup.Stderr)
	}
	if resp.SendSecond.ExitCode != 0 {
		t.Fatalf("second send exit=%d: %s", resp.SendSecond.ExitCode, resp.SendSecond.Stderr)
	}
	if !resp.HasHello {
		t.Fatalf("want hello marker; snap=\n%s\nevents=\n%s", resp.ResumeSnapshot, resp.EventsBlob)
	}
}
```
