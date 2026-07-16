## Expected Output

Stderr plans the add-all step:

```text
<contains>
would: git add -A
</contains>
```

## Expected
- Stderr contains the dry-run plan line `would: git add -A`.
- Index is **unchanged**: the untracked file is still not staged.
- With an empty index, a non-zero exit with `no staged` error after the would-line
  is OK (honest dry-run counts the current index only).
- Agent is not invoked.

## Side Effects
- No `git add` mutation: `git diff --cached --name-only` does not list the untracked file.
- HEAD subject unchanged.

## Exit Code
- Non-zero when the current index is empty (expected honest dry-run error), or
  zero only if implementation somehow still succeeds without staging (must not).

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	name := req.Operation
	if name == "" {
		name = AddAllUntrackedName
	}

	if !strings.Contains(resp.Stderr, "would: git add -A") {
		t.Fatalf("stderr must contain would: git add -A, got stderr:\n%s\nstdout:\n%s\nerr: %v",
			resp.Stderr, resp.Stdout, resp.Err)
	}
	// Real add must not appear as executed under dry-run.
	if strings.Contains(resp.Stderr, "$ git add -A") {
		t.Fatalf("dry-run must not log real $ git add -A, stderr:\n%s", resp.Stderr)
	}

	staged := GitStagedNamesAddAll(t, req.GitDir)
	for _, s := range staged {
		if s == name {
			t.Fatalf("untracked %q must remain unstaged under --add-all --dry-run, staged=%v", name, staged)
		}
	}

	// Empty-index honest dry-run: error after would-line is expected.
	if resp.Err == nil {
		t.Fatalf("empty index under --add-all --dry-run should still error no staged (index not mutated), stdout:\n%s\nstderr:\n%s",
			resp.Stdout, resp.Stderr)
	}
	errMsg := strings.ToLower(resp.Err.Error())
	if !strings.Contains(errMsg, "no staged") {
		t.Fatalf("error should mention no staged changes after would-line, got: %v\nstderr:\n%s",
			resp.Err, resp.Stderr)
	}

	for _, marker := range []string{"Passing diff to agent", "Running agent"} {
		if strings.Contains(resp.Stderr, marker) {
			t.Fatalf("agent must not run under dry-run, found %q in stderr:\n%s", marker, resp.Stderr)
		}
	}

	assert.Output(t, resp.Stderr, `<contains>
would: git add -A
</contains>
`)
}
```
