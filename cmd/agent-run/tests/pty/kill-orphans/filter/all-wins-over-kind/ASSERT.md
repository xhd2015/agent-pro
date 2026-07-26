---
label: e2e
---

## Expected

- Exit code 0.
- Stdout lists the non-TestGenerated child PID despite `--kind=test-generated`
  (because `--all` wins).
- Serve still alive; trailing `\n`.

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
		t.Fatalf("--all must win over --kind and list pid %s; stdout:\n%s", pidStr, resp.Stdout)
	}
	if !processAlive(pid) {
		t.Fatalf("dry-run must not kill serve pid %d", pid)
	}
	assertTrailingNewline(t, resp.Stdout, "all-wins-over-kind dry-run stdout")
}
```
