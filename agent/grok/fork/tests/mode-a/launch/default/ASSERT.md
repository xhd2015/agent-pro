## Expected Output

```
Opened new window; launching grok-fork --session-id 019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa
```

Output ends with a newline.

## Expected

- Exit 0.
- Exactly one `OpenInNewTerminal`; dir = session cwd; follow-up is executable + `--session-id`.
- Follow-up does not contain `grok --resume`.
- No ANSI (auto-on-pipe).
- No RunForeground.

## Side Effects

- Recorded open only.

## Errors

- None.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainOK(t, resp)
	assertNoForeground(t, resp)
	assertNoANSI(t, resp.Stdout, "success stdout")
	assertStdoutExact(t, resp.Stdout, modeASuccessLine(fixtureSessionID))
	call := assertOneOpen(t, resp)
	if call.Dir != req.Workspace {
		t.Fatalf("open dir: got %q, want %q", call.Dir, req.Workspace)
	}
	want := followUpSession(req.Executable, fixtureSessionID)
	if call.FollowUp != want {
		t.Fatalf("follow-up: got %q, want %q", call.FollowUp, want)
	}
	if strings.Contains(call.FollowUp, "grok --resume") {
		t.Fatalf("follow-up must not contain grok --resume: %q", call.FollowUp)
	}
}
```
