## Expected Output

Stdout includes the generated title:

```text
<contains>
feat: via commandcode
</contains>
```

## Expected
- gen-commit-msg `--commit` succeeds with commandcode mock.
- A new commit is created with subject `feat: via commandcode`.
- HEAD subject differs from pre-run subject (`initial`).

## Side Effects
- Git log shows the mock-generated commit title as the latest subject.

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
		t.Fatalf("gen-commit-msg --commit with commandcode should succeed, got: %v\nstdout:\n%s\nstderr:\n%s",
			resp.Err, resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "feat: via commandcode") {
		t.Fatalf("stdout missing title, got:\n%s", resp.Stdout)
	}
	got := GitHEADSubjectCmd(t, req.GitDir)
	if got != "feat: via commandcode" {
		t.Fatalf("commit subject = %q, want %q (before was %q)", got, "feat: via commandcode", req.HEADSubjectBefore)
	}
	if got == req.HEADSubjectBefore {
		t.Fatalf("HEAD subject should change after --commit, still %q", got)
	}
	assert.Output(t, resp.Stdout, `<contains>
feat: via commandcode
</contains>
`)
}
```
