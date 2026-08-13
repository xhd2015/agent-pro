## Expected

- Exit 0.
- One `RunForeground`: bin basename `llm-mock-run-grok`; argv `--resume <id> --fork-session`; dir = session cwd.
- No `OpenInNewTerminal`.

## Side Effects

- Recorded foreground only.

## Errors

- None.

## Exit Code

0

```go
import (
	"path/filepath"
	"reflect"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainOK(t, resp)
	assertNoOpen(t, resp)
	call := assertOneForeground(t, resp)
	if filepath.Base(call.Bin) != "llm-mock-run-grok" {
		t.Fatalf("grok bin basename: got %q, want llm-mock-run-grok (bin=%q)", filepath.Base(call.Bin), call.Bin)
	}
	wantArgv := []string{"--resume", fixtureSessionID, "--fork-session"}
	if !reflect.DeepEqual(call.Argv, wantArgv) {
		t.Fatalf("argv: got %#v, want %#v", call.Argv, wantArgv)
	}
	if call.Dir != req.Workspace {
		t.Fatalf("foreground dir: got %q, want %q", call.Dir, req.Workspace)
	}
}
```
