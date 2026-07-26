## Expected Output

Stdout includes the mock-generated title:

```text
<contains>
feat: commit untracked
</contains>
```

## Expected
- gen-commit-msg succeeds with `--add-all --commit`.
- Stderr contains `$ git add -A`.
- HEAD subject becomes `feat: commit untracked` (differs from pre-run subject).
- The previously untracked file is present in the new commit.

## Side Effects
- New commit on HEAD containing `untracked.go`.

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
		// Classic TDD RED until --add-all is implemented.
		t.Fatalf("gen-commit-msg --add-all --commit should succeed, got: %v\nstdout:\n%s\nstderr:\n%s",
			resp.Err, resp.Stdout, resp.Stderr)
	}

	if !strings.Contains(resp.Stderr, "$ git add -A") {
		t.Fatalf("stderr must contain $ git add -A, stderr:\n%s", resp.Stderr)
	}

	wantSubject := "feat: commit untracked"
	gotHEAD := GitHEADSubjectAddAll(t, req.GitDir)
	if gotHEAD != wantSubject {
		t.Fatalf("commit subject = %q, want %q (before was %q)\nstdout:\n%s\nstderr:\n%s",
			gotHEAD, wantSubject, req.HEADSubjectBefore, resp.Stdout, resp.Stderr)
	}
	if gotHEAD == req.HEADSubjectBefore {
		t.Fatalf("HEAD subject should change after --add-all --commit, still %q", gotHEAD)
	}

	name := req.Operation
	if name == "" {
		name = AddAllUntrackedName
	}
	files := GitHEADFilesAddAll(t, req.GitDir)
	found := false
	for _, f := range files {
		if f == name {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("new commit must include %q, files=%v", name, files)
	}

	if !strings.Contains(resp.Stdout, wantSubject) {
		t.Fatalf("stdout missing title, got:\n%s", resp.Stdout)
	}

	assert.Output(t, resp.Stdout, `<contains>
feat: commit untracked
</contains>
`)
}
```
