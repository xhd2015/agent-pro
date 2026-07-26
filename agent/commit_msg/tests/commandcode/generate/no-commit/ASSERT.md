## Expected Output

Stdout includes the formatted title and description from the mock JSON:

```text
<contains>
feat: via commandcode
</contains>
```

## Expected
- gen-commit-msg succeeds with `--agent-runner commandcode` and mock binary.
- stdout contains title `feat: via commandcode`.
- stdout contains description `from commandcode mock`.
- HEAD subject is unchanged (no `--commit`).

## Side Effects
- No new git commit.

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
	if resp.Err != nil {
		// Today RED until commandcode runner is implemented.
		t.Fatalf("gen-commit-msg commandcode generate should succeed, got: %v\nstdout:\n%s\nstderr:\n%s",
			resp.Err, resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "feat: via commandcode") {
		t.Fatalf("stdout missing title, got:\n%s", resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, "from commandcode mock") {
		t.Fatalf("stdout missing description, got:\n%s", resp.Stdout)
	}
	gotHEAD := GitHEADSubjectCmd(t, req.GitDir)
	if gotHEAD != req.HEADSubjectBefore {
		t.Fatalf("HEAD subject changed without --commit: before=%q after=%q",
			req.HEADSubjectBefore, gotHEAD)
	}
	assert.Output(t, resp.Stdout, `<contains>
feat: via commandcode
</contains>
`)
}
```
