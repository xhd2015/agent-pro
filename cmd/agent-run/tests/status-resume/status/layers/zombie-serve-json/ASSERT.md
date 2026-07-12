## Expected

- Exit code 0.
- JSON `process.status` is `alive` (serve keep-alive).
- JSON `terminal.status` is `reachable`; `terminal.sendable` is `no` when present.
- JSON `runner.status` is `bound`; `runner.exited` is boolean `true`.
- JSON `resume.ready` is boolean `true`.
- Stdout ends with trailing `\n`.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	if resp.JSONBody == nil {
		t.Fatalf("expected parsed JSON body, stdout:\n%s", resp.Stdout)
	}
	body := resp.JSONBody

	if st, ok := jsonPathString(body, "process", "status"); !ok || !strings.EqualFold(st, "alive") {
		t.Fatalf("process.status = %q, want alive; body=%v", st, body)
	}
	if st, ok := jsonPathString(body, "terminal", "status"); !ok || !strings.EqualFold(st, "reachable") {
		t.Fatalf("terminal.status = %q, want reachable; body=%v", st, body)
	}
	if s, ok := jsonPathString(body, "terminal", "sendable"); ok && !strings.EqualFold(s, "no") {
		t.Fatalf("terminal.sendable = %q, want no; body=%v", s, body)
	}
	if st, ok := jsonPathString(body, "runner", "status"); !ok || !strings.EqualFold(st, "bound") {
		t.Fatalf("runner.status = %q, want bound; body=%v", st, body)
	}
	if exited, ok := jsonPathBool(body, "runner", "exited"); ok {
		if !exited {
			t.Fatalf("runner.exited = false, want true (zombie serve after /exit); body=%v", body)
		}
	} else if s, ok := jsonPathString(body, "runner", "exited"); ok {
		if s != "true" {
			t.Fatalf("runner.exited = %q, want true; body=%v", s, body)
		}
	} else {
		t.Fatalf("JSON missing runner.exited; body=%v", body)
	}
	if ready, ok := jsonPathBool(body, "resume", "ready"); ok {
		if !ready {
			t.Fatalf("resume.ready = false, want true; body=%v", body)
		}
	} else if s, ok := jsonPathString(body, "resume", "ready"); ok {
		low := strings.ToLower(s)
		if low != "true" && low != "yes" {
			t.Fatalf("resume.ready = %q, want true/yes; body=%v", s, body)
		}
	} else {
		t.Fatalf("JSON missing resume.ready; body=%v", body)
	}
	assertTrailingNewline(t, resp.Stdout, "status --json stdout")
}
```
