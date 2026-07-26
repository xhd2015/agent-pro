## Expected Output

Stdout is exactly mock message B for N=2, ending with a trailing newline:

```text
dry-run: would generate commit message for 2 staged file(s)
```

## Expected
- gen-commit-msg exits successfully with `--dry-run`.
- stdout is the exact mock message for 2 staged files (trailing newline).
- Agent is not invoked (no agent-phase stderr markers).

## Side Effects
- Index is unchanged (both text files remain staged).
- No commit is created.

## Exit Code
- Zero.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("dry-run should succeed, got: %v\nstderr:\n%s", resp.Err, resp.Stderr)
	}
	AssertMockMessageB(t, resp.Stdout, 2)
	AssertNoAgentInvoked(t, resp)

	staged := GitStagedNames(t, req.GitDir)
	joined := strings.Join(staged, "\n")
	for _, want := range []string{"change_1.go", "change_2.go"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected %q still staged, staged=%v", want, staged)
		}
	}
}
```
