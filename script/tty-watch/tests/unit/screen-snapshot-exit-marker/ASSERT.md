## Expected

- Scrollback `yes\n\n[Terminal exited]` (LF+LF before exit marker, no CR) converts to
  exactly two left-aligned lines: `yes` and `[Terminal exited]`.
- Snapshot text must NOT start with `\n` (no leading blank line before `yes`).
- No leading spaces before `[Terminal exited]` in snapshot text (vt10x column drift).

```go
import (
	"reflect"
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/script/tty-watch/ttywatchtest"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.SnapshotText == "" {
		t.Fatal("expected snapshot text in response")
	}
	if err := ttywatchtest.NoLeadingBlankLine(resp.SnapshotText); err != nil {
		t.Fatalf("snapshot text must not start with blank line: %v; raw %q", err, resp.SnapshotText)
	}
	if strings.HasPrefix(resp.SnapshotText, "\n") {
		t.Fatalf("snapshot text must not start with \\n; raw %q", resp.SnapshotText)
	}
	lines, err := ttywatchtest.ContentLinesAtColumnZero(resp.SnapshotText)
	if err != nil {
		t.Fatalf("snapshot text not left-aligned: %v; raw %q", err, resp.SnapshotText)
	}
	want := []string{"yes", "[Terminal exited]"}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("snapshot lines = %v, want %v; raw %q", lines, want, resp.SnapshotText)
	}
	if strings.Contains(resp.SnapshotText, "   [Terminal exited]") {
		t.Fatalf("exit marker indented in snapshot text: %q", resp.SnapshotText)
	}
}
```