## Expected Output

Stdout is exact mock message B for N=1:

```text
dry-run: would generate commit message for 1 staged file(s)
```

## Expected
- gen-commit-msg accepts `--agent-runner commandcode` under `--dry-run` (not unsupported).
- Exit success; stdout is mock B for 1 staged file.
- Agent is not invoked (no agent-phase stderr markers).

## Side Effects
- Index unchanged; no commit.

## Exit Code
- Zero.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.Err != nil {
		// Today RED: commandcode is rejected as unsupported until implementer lands.
		t.Fatalf("dry-run --agent-runner commandcode should succeed, got: %v\nstdout:\n%s\nstderr:\n%s",
			resp.Err, resp.Stdout, resp.Stderr)
	}
	want := "dry-run: would generate commit message for 1 staged file(s)\n"
	if resp.Stdout != want {
		t.Fatalf("stdout mock B mismatch\n got: %q\nwant: %q", resp.Stdout, want)
	}
	for _, marker := range []string{"Passing diff to agent", "Running agent", "agent failed"} {
		if strings.Contains(resp.Stderr, marker) {
			t.Fatalf("agent must not run under --dry-run, found %q in stderr:\n%s", marker, resp.Stderr)
		}
	}
}
```
