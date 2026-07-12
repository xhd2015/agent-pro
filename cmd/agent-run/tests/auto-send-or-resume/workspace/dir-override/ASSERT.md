## Expected

- Exit code 0.
- Cwd probe equals `--dir` override path (not meta.workspace, not CLI cwd).
- Argv includes `--resume`.

## Exit Code

0

```go
import (
	"os"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v\nstdout:\n%s\nstderr:\n%s", err, resp.Stdout, resp.Stderr)
	}
	assertSuccess(t, resp)

	cwd := resp.CwdProbe
	if cwd == "" {
		b, rErr := os.ReadFile(req.CwdProbePath)
		if rErr != nil {
			t.Fatalf("read cwd probe %s: %v\nstderr:\n%s", req.CwdProbePath, rErr, resp.Stderr)
		}
		cwd = strings.TrimSpace(string(b))
	}
	want := canonicalPath(t, req.DirOverride)
	got := canonicalPath(t, cwd)
	if got != want {
		t.Fatalf("child cwd must be --dir override\n  got:  %s\n  want: %s\n  meta.workspace: %s\n  cli: %s\nstderr:\n%s",
			got, want, canonicalPath(t, req.Workspace), canonicalPath(t, req.WorkDir), resp.Stderr)
	}
	// Must not accidentally use meta.workspace when --dir set.
	if got == canonicalPath(t, req.Workspace) && want != canonicalPath(t, req.Workspace) {
		t.Fatal("--dir must win over meta.workspace")
	}

	probe := resp.ArgvProbe
	if probe == "" {
		b, _ := os.ReadFile(req.ArgvProbePath)
		probe = string(b)
	}
	assertContains(t, probe, "--resume")
	assertContains(t, probe, req.RunnerSessionID)
}
```
