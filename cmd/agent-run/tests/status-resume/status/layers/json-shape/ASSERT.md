## Expected

- Exit code 0.
- Stdout is a JSON object with nested `runner` and `resume` objects.
- `runner.exited` is boolean `true` (or null only if unknown — for this seed: true).
- `resume.ready` is boolean `true` or string/bool equivalent yes.
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
	// runner.exited
	if exited, ok := jsonPathBool(resp.JSONBody, "runner", "exited"); ok {
		if !exited {
			t.Fatalf("runner.exited = false, want true; body=%v", resp.JSONBody)
		}
	} else if s, ok := jsonPathString(resp.JSONBody, "runner", "exited"); ok {
		if s != "true" {
			t.Fatalf("runner.exited = %q, want true; body=%v", s, resp.JSONBody)
		}
	} else {
		t.Fatalf("JSON missing runner.exited; body=%v", resp.JSONBody)
	}
	// resume.ready
	if ready, ok := jsonPathBool(resp.JSONBody, "resume", "ready"); ok {
		if !ready {
			t.Fatalf("resume.ready = false, want true; body=%v", resp.JSONBody)
		}
	} else if s, ok := jsonPathString(resp.JSONBody, "resume", "ready"); ok {
		low := strings.ToLower(s)
		if low != "true" && low != "yes" {
			t.Fatalf("resume.ready = %q, want true/yes; body=%v", s, resp.JSONBody)
		}
	} else {
		t.Fatalf("JSON missing resume.ready; body=%v", resp.JSONBody)
	}
	assertTrailingNewline(t, resp.Stdout, "status --json stdout")
}
```
