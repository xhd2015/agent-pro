## Expected

- Exit 0.
- Hook marker file exists under isolated `GROK_HOME`.
- Grok bin basename is `llm-mock-run-grok`.
- No `OpenInNewTerminal`.

## Side Effects

- Marker file `$GROK_HOME/fork-hook-ok.txt` created by the child hook.

## Errors

- None.

## Exit Code

0

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainOK(t, resp)
	assertNoOpen(t, resp)
	if req.HookMarker == "" {
		t.Fatal("HookMarker empty")
	}
	if _, statErr := os.Stat(req.HookMarker); statErr != nil {
		t.Fatalf("hook marker missing %s: %v", req.HookMarker, statErr)
	}
	if len(resp.ForegroundCalls) != 1 {
		t.Fatalf("expected one real foreground exec, got %d", len(resp.ForegroundCalls))
	}
	if filepath.Base(resp.ForegroundCalls[0].Bin) != "llm-mock-run-grok" {
		t.Fatalf("exec bin basename: %q", resp.ForegroundCalls[0].Bin)
	}
}
```
