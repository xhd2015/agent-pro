## Expected Output

Help text documents supported agent runners including `commandcode`:

```text
<contains>
commandcode
</contains>
```

## Expected
- Exit code 0 (CLI help path).
- Combined stdout/stderr is non-empty usage text.
- Help mentions `commandcode` as a supported agent runner.
- Help still mentions `opencode` (default remains opencode).
- Help mentions `--agent-runner`.

## Side Effects
- Read-only (`-h` only); no git mutation.

## Exit Code
- Zero.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("help should exit 0, got %d\nstdout:\n%s\nstderr:\n%s\nerr: %v",
			resp.ExitCode, resp.Stdout, resp.Stderr, resp.Err)
	}
	help := resp.Stdout + resp.Stderr
	if strings.TrimSpace(help) == "" {
		t.Fatal("expected non-empty help text")
	}
	if !strings.Contains(help, "commandcode") {
		t.Fatalf("help must mention commandcode as supported agent runner; got:\n%s", help)
	}
	if !strings.Contains(help, "opencode") {
		t.Fatalf("help must still mention opencode; got:\n%s", help)
	}
	if !strings.Contains(help, "--agent-runner") {
		t.Fatalf("help must mention --agent-runner; got:\n%s", help)
	}
	assert.Output(t, help, `<contains>
commandcode
</contains>
`)
}
```
