## Expected

- Exit code 0.
- COMMAND cell for the session ends with `...`.
- COMMAND cell display width is at most 64 characters (61-char prefix + `...` when truncated).

```go
import (
	"strings"
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
	fullCmd := strings.Join(req.RunCommand, " ")
	cmd, ok := ttywatchtest.ListTableCommand(resp.ListOutput, resp.SessionID)
	if !ok {
		t.Fatalf("list missing parseable COMMAND cell for %s: %q", resp.SessionID, resp.ListOutput)
	}
	if len(cmd) > 64 {
		t.Fatalf("COMMAND cell len %d exceeds 64: %q", len(cmd), cmd)
	}
	if !strings.HasSuffix(cmd, "...") {
		t.Fatalf("truncated COMMAND should end with ...: %q", cmd)
	}
	want := fullCmd
	if len(fullCmd) > 64 {
		want = fullCmd[:61] + "..."
	}
	if cmd != want {
		t.Fatalf("COMMAND cell = %q, want %q (full joined argv %q)", cmd, want, fullCmd)
	}
}
```