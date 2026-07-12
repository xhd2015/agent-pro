## Expected

- Primary exit 0; stdout lists non-PPID1 child PID when `--all` is set.
- Follow-up default dry-run does **not** list that PID.
- Serve still alive; primary stdout trailing `\n`.

## Exit Code

0

```go
import (
	"strconv"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	pid := 0
	if resp.SpawnPIDs != nil {
		pid = resp.SpawnPIDs["child"]
	}
	if pid <= 0 {
		pid = resp.ServePID
	}
	if pid <= 0 {
		t.Fatalf("expected child serve PID")
	}
	pidStr := strconv.Itoa(pid)
	if !strings.Contains(resp.Stdout, pidStr) {
		t.Fatalf("--all dry-run must list non-PPID1 pid %s; stdout:\n%s", pidStr, resp.Stdout)
	}
	if resp.FollowUpExitCode != 0 {
		t.Fatalf("follow-up default dry-run exit %d; stderr:\n%s",
			resp.FollowUpExitCode, resp.FollowUpStderr)
	}
	if strings.Contains(resp.FollowUpStdout, pidStr) {
		t.Fatalf("default dry-run must NOT list pid %s; follow-up:\n%s", pidStr, resp.FollowUpStdout)
	}
	if !processAlive(pid) {
		t.Fatalf("dry-run must not kill serve pid %d", pid)
	}
	assertTrailingNewline(t, resp.Stdout, "--all dry-run stdout")
}
```
