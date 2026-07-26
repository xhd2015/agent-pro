## Expected Output

Stdout includes the mock-generated title:

```text
<contains>
feat: add untracked
</contains>
```

## Expected
- gen-commit-msg succeeds with `--add-all` and no `--commit`.
- Stderr contains the real log line `$ git add -A`.
- The previously untracked file is staged after the run (or generation succeeded
  only because the staged set became non-empty).
- Stdout contains title `feat: add untracked`.
- HEAD subject is unchanged (no `--commit`).

## Side Effects
- Index includes `untracked.go` after the run.
- No new commit.

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
		// Classic TDD RED until --add-all is implemented (unknown flag or missing stage).
		t.Fatalf("gen-commit-msg --add-all should succeed and stage untracked, got: %v\nstdout:\n%s\nstderr:\n%s",
			resp.Err, resp.Stdout, resp.Stderr)
	}

	if !strings.Contains(resp.Stderr, "$ git add -A") {
		t.Fatalf("stderr must contain $ git add -A, stderr:\n%s", resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "would: git add -A") {
		t.Fatalf("real path must not print dry-run would: git add -A, stderr:\n%s", resp.Stderr)
	}

	name := req.Operation
	if name == "" {
		name = AddAllUntrackedName
	}
	staged := GitStagedNamesAddAll(t, req.GitDir)
	found := false
	for _, s := range staged {
		if s == name {
			found = true
			break
		}
	}
	if !found {
		// Generation success alone is insufficient if file never landed in index.
		t.Fatalf("expected %q staged after --add-all, staged=%v\nstdout:\n%s\nstderr:\n%s",
			name, staged, resp.Stdout, resp.Stderr)
	}

	if !strings.Contains(resp.Stdout, "feat: add untracked") {
		t.Fatalf("stdout missing title, got:\n%s", resp.Stdout)
	}

	gotHEAD := GitHEADSubjectAddAll(t, req.GitDir)
	if gotHEAD != req.HEADSubjectBefore {
		t.Fatalf("HEAD subject changed without --commit: before=%q after=%q",
			req.HEADSubjectBefore, gotHEAD)
	}

	assert.Output(t, resp.Stdout, `<contains>
feat: add untracked
</contains>
`)
}
```
