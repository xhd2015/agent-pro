## Expected

- Exit code 0.
- Data row for the session shows `WATCH=0` and `ATTACHED=0`.

```go
import (
	"testing"

	"github.com/xhd2015/agent-pro/script/tty-watch/ttywatchtest"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("list exit %d, output %q", resp.ExitCode, resp.ListOutput)
	}
	if resp.SessionID == "" {
		t.Fatal("expected session id from harness setup")
	}
	watch, attached, ok := ttywatchtest.ListTableClientCounts(resp.ListOutput, resp.SessionID)
	if !ok {
		t.Fatalf("list missing parseable table row for %s: %q", resp.SessionID, resp.ListOutput)
	}
	if watch != 0 {
		t.Fatalf("expected WATCH=0, got %d: %q", watch, resp.ListOutput)
	}
	if attached != 0 {
		t.Fatalf("expected ATTACHED=0, got %d: %q", attached, resp.ListOutput)
	}
}
```