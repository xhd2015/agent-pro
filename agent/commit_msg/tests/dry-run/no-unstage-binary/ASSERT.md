## Expected
- Dry-run succeeds.
- Stdout mock B uses N=2 (binary + text, count before unstage).
- Stderr mentions a planned unstage of the binary (`would` + `unstage` + binary path).
- Binary remains staged after the run (no index mutation).

## Side Effects
- `git diff --cached --name-only` still lists the binary and the text file.
- Agent is not invoked.

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
		t.Fatalf("dry-run with staged binary should succeed, got: %v\nstderr:\n%s", resp.Err, resp.Stderr)
	}
	AssertMockMessageB(t, resp.Stdout, 2)
	AssertNoAgentInvoked(t, resp)

	binRel := req.Operation
	if binRel == "" {
		binRel = "blob.bin"
	}
	stderrLower := strings.ToLower(resp.Stderr)
	if !strings.Contains(stderrLower, "would") || !strings.Contains(stderrLower, "unstage") {
		t.Fatalf("stderr should plan unstage with would/unstage, stderr:\n%s", resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, binRel) {
		t.Fatalf("stderr should mention binary %q, stderr:\n%s", binRel, resp.Stderr)
	}

	staged := GitStagedNames(t, req.GitDir)
	joined := strings.Join(staged, "\n")
	if !strings.Contains(joined, binRel) {
		t.Fatalf("binary %q must remain staged after dry-run, staged=%v", binRel, staged)
	}
	if !strings.Contains(joined, "app.go") {
		t.Fatalf("text file app.go must remain staged, staged=%v", staged)
	}
}
```
