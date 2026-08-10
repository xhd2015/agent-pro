## Expected

- Exit code 0.
- Stdout (or combined) contains dry-run plan for kill of pid `9202`.
- Stdout contains dry-run plan for opening iTerm2 (follow-up command mentioned).
- Kill log empty (no signals recorded).
- No iTerm script content written.
- No agent-run session meta created.

## Side Effects

- No kill; no meta create; no iTerm open.

## Exit Code

0

```go
import (
	"strconv"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	const nativePID = 9202
	combined := combinedOut(resp)
	assertTakeoverActionImplemented(t, combined)
	assertExitCode(t, resp, 0)

	lower := strings.ToLower(combined)
	if !strings.Contains(lower, "dry-run") && !strings.Contains(lower, "dry run") {
		t.Fatalf("expected dry-run plan text, got:\n%s", combined)
	}
	if !strings.Contains(combined, strconv.Itoa(nativePID)) {
		t.Fatalf("dry-run plan must mention pid %d, got:\n%s", nativePID, combined)
	}
	assertContainsAny(t, combined, "kill", "would kill")
	assertContainsAny(t, combined, "iterm", "iTerm", "open")

	assertNoKillLog(t, req)
	assertNoItermScript(t, req)

	if ids := listAgentSessionIDs(t, req.Home); len(ids) > 0 {
		t.Fatalf("dry-run must not create agent-run meta; sessions=%v", ids)
	}
}
```
