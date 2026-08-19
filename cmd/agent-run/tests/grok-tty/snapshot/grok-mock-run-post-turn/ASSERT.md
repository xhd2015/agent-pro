---
label: e2e
---

## Expected

- `agent-run snapshot` exits 0.
- Snapshot shows the submitted user prompt in the grok TUI screen (parity with
  `tty-watch snapshot` on an equivalent llm-mock grok session).
- Snapshot is not limited to the post-turn status bar only (`Turn completed`,
  `Ctrl+.:shortcuts`) without conversation/UI content.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.SnapshotExitCode != 0 {
		t.Fatalf("snapshot exit %d stdout %q stderr %q", resp.SnapshotExitCode, resp.SnapshotStdout, resp.Stderr)
	}
	text := strings.TrimSpace(resp.SnapshotStdout)
	if text == "" {
		t.Fatalf("snapshot stdout empty; background stdout:\n%s\nbackground stderr:\n%s",
			resp.BackgroundStdout, resp.BackgroundStderr)
	}
	// Substring check: PTY snapshots may glue adjacent printf rows
	// ("Grok Build Beta" + "one word…") so line-oriented assert.Output flakes.
	if !strings.Contains(text, "one word of France captial") {
		t.Fatalf("snapshot missing prompt; got %q", text)
	}
	if !strings.Contains(text, "New worktree") && !strings.Contains(text, "Grok Build") {
		t.Fatalf("snapshot missing grok TUI screen content (want menu or build banner like tty-watch snapshot), got %q", text)
	}
	if strings.Contains(text, "Turn completed") && !strings.Contains(text, "one word of France captial") {
		t.Fatalf("snapshot shows post-turn status bar without conversation prompt (tty-watch snapshot includes prompt), got %q", text)
	}
}
```