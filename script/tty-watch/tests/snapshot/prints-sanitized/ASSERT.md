## Expected

- Exit code 0.
- Snapshot text contains `PLAIN_LINE`.
- Snapshot text contains readable `RED` label (ANSI stripped).
- `resp.ContainsEscape` is false (no CSI/OSC/C0 controls except \n \r \t).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("snapshot exit %d stdout %q stderr %q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	text := resp.SnapshotText
	if text == "" {
		text = resp.Stdout
	}
	if !strings.Contains(text, "PLAIN_LINE") {
		t.Fatalf("snapshot missing PLAIN_LINE: %q", text)
	}
	if !strings.Contains(text, "RED") {
		t.Fatalf("snapshot missing RED label: %q", text)
	}
	if resp.ContainsEscape {
		t.Fatalf("snapshot still contains escape sequences: %q", text)
	}
}
```