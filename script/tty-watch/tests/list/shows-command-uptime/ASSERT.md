## Expected

- Exit code 0.
- List output contains session id (`session-N`).
- List output contains command name `sleep`.
- List output contains an uptime indicator (e.g. `s`, `m`, `h`, `ago`, or digit).

```go
import (
	"regexp"
	"strings"
	"testing"
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
	if !strings.Contains(out, resp.SessionID) {
		t.Fatalf("list missing session id %s: %q", resp.SessionID, out)
	}
	if !strings.Contains(out, "sleep") {
		t.Fatalf("list missing command sleep: %q", out)
	}
	uptimePattern := regexp.MustCompile(`(?i)(\d+\s*(s|sec|second|m|min|h|hour)|ago|uptime)`)
	if !uptimePattern.MatchString(out) {
		t.Fatalf("list missing uptime hint: %q", out)
	}
}
```