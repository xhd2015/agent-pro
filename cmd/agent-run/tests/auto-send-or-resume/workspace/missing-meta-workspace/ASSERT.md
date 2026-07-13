## Expected

- Exit code **1** (non-zero).
- Stderr (or stdout) identifies the problem as **session workspace missing / no longer exists** and includes the missing path (`req.Workspace` / `gone-ws`).
- Stderr hints the user to pass **`--dir`** to override with an existing directory.
- Failure must **not** be only the misleading binary-missing shape (`fork/exec … no such file or directory`) without workspace context.
- Provider must not silently fall back to process cwd (exit 0 + wrong cwd).

## Errors

- Clear workspace-missing error naming the path.
- Hint mentions `--dir`.

## Exit Code

1

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)

	combined := resp.Stderr + "\n" + resp.Stdout
	lower := strings.ToLower(combined)

	// Path must appear (full or basename) so the user knows which workspace is gone.
	pathHit := strings.Contains(combined, req.Workspace) ||
		strings.Contains(combined, filepath.Base(req.Workspace))
	if !pathHit {
		t.Fatalf("error must include missing workspace path %q (or basename)\ncombined:\n%s",
			req.Workspace, combined)
	}

	// Name the problem as workspace-related (not only chdir/fork/exec binary noise).
	assertContainsAny(t, lower,
		"workspace",
		"session workspace",
		"meta.workspace",
	)
	assertContainsAny(t, lower,
		"does not exist",
		"do not exist",
		"no longer exist",
		"no longer exists",
		"missing",
		"not found",
		"not a directory",
		"no such file or directory",
	)

	// Override hint.
	assertContainsAny(t, lower,
		"--dir",
		"pass --dir",
		"use --dir",
		"provide --dir",
	)

	// Misleading binary-only fork/exec without workspace context must fail the
	// workspace assert above; keep an explicit guard for reviewers.
	if strings.Contains(lower, "fork/exec") && !strings.Contains(lower, "workspace") {
		t.Fatalf("fork/exec failure must still name session workspace problem\ncombined:\n%s", combined)
	}
}
```
