---
label: e2e
---

## Expected

- Exit code **1** (non-zero) on auto→resume when `meta.workspace` is missing and `--dir` is unset.
- Stderr (or stdout) identifies **session workspace missing / no longer exists** and includes the missing path.
- Stderr hints **`--dir`** as the override.
- Must not present failure as missing `agent-run` binary only (`fork/exec` without workspace context).
- Must not silently fall back to process cwd (no exit 0 spawn).

## Errors

- Same user-facing contract as `workspace/missing-meta-workspace` (resume subcommand).

## Exit Code

1

```go
import (
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertExitCode(t, resp, 1)

	combined := resp.Stderr + "\n" + resp.Stdout
	lower := strings.ToLower(combined)

	pathHit := strings.Contains(combined, req.Workspace) ||
		strings.Contains(combined, filepath.Base(req.Workspace))
	if !pathHit {
		t.Fatalf("auto→resume error must include missing workspace path %q (or basename)\ncombined:\n%s",
			req.Workspace, combined)
	}

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
	assertContainsAny(t, lower,
		"--dir",
		"pass --dir",
		"use --dir",
		"provide --dir",
	)

	if strings.Contains(lower, "fork/exec") && !strings.Contains(lower, "workspace") {
		t.Fatalf("fork/exec failure must still name session workspace problem\ncombined:\n%s", combined)
	}
}
```
