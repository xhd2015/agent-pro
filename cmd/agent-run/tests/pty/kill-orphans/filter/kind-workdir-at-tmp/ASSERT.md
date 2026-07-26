---
label: e2e
---

## Expected

- Primary exit code 0; stdout lists the workdir-at-tmp child PID.
- Follow-up default dry-run does **not** list that PID.
- Serve remains alive; primary stdout trailing `\n`.

## Exit Code

0

```go
import (
	"strconv"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	pid := 0
	if resp.SpawnPIDs != nil {
		pid = resp.SpawnPIDs["tmp"]
	}
	if pid <= 0 {
		pid = resp.ServePID
	}
	if pid <= 0 {
		t.Fatalf("expected workdir-at-tmp serve PID")
	}
	pidStr := strconv.Itoa(pid)
	if !strings.Contains(resp.Stdout, pidStr) {
		t.Fatalf("--kind=workdir-at-tmp dry-run must list pid %s; stdout:\n%s", pidStr, resp.Stdout)
	}
	if resp.FollowUpExitCode != 0 {
		t.Fatalf("follow-up default dry-run exit %d; stderr:\n%s",
			resp.FollowUpExitCode, resp.FollowUpStderr)
	}
	if strings.Contains(resp.FollowUpStdout, pidStr) {
		t.Fatalf("default dry-run must NOT list PPID≠1 workdir pid %s; follow-up:\n%s",
			pidStr, resp.FollowUpStdout)
	}
	if !processAlive(pid) {
		t.Fatalf("dry-run must not kill serve pid %d", pid)
	}
	assertTrailingNewline(t, resp.Stdout, "kind-workdir-at-tmp dry-run stdout")
}
```
