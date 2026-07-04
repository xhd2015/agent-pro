## Expected

- Exit code 0.
- List output includes header columns in order: `SESSION`, `UPTIME`, `WATCH`, `ATTACHED`, `COMMAND` (COMMAND last).
- Header and first data row columns align (WATCH/ATTACHED numeric fields line up under headers).

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
	out := resp.ListOutput
	if !ttywatchtest.ListTableHeaderPresent(out) {
		t.Fatalf("list missing table header columns: %q", out)
	}
	if !ttywatchtest.ListTableColumnsAligned(out) {
		t.Fatalf("list header and data row columns not aligned: %q", out)
	}
	watch, attached, ok := ttywatchtest.ListTableClientCounts(out, resp.SessionID)
	if !ok {
		t.Fatalf("list missing parseable table row for %s: %q", resp.SessionID, out)
	}
	_ = watch
	_ = attached
}
```