---
label: e2e
---

## Expected

- Exit code 0.
- Stdout lists the PPID1 orphan PID (label `ppid1`).
- Stdout does **not** list the non-orphan child PID (label `child`).
- Stdout ends with trailing newline `\n`.
- Dry-run leaves both processes alive.

## Side Effects

- Neither serve is killed by dry-run; harness cleanup terminates them.

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
	if resp.SpawnPIDs == nil {
		t.Fatalf("expected SpawnPIDs for dual spawn")
	}
	ppid1 := resp.SpawnPIDs["ppid1"]
	child := resp.SpawnPIDs["child"]
	if ppid1 <= 0 || child <= 0 {
		t.Fatalf("expected ppid1 and child PIDs; got %#v", resp.SpawnPIDs)
	}
	out := resp.Stdout
	if !strings.Contains(out, strconv.Itoa(ppid1)) {
		t.Fatalf("default dry-run must list PPID1 orphan pid %d; stdout:\n%s", ppid1, out)
	}
	if strings.Contains(out, strconv.Itoa(child)) {
		t.Fatalf("default dry-run must NOT list non-orphan child pid %d; stdout:\n%s", child, out)
	}
	if !processAlive(ppid1) || !processAlive(child) {
		t.Fatalf("dry-run must not kill serves; ppid1 alive=%v child alive=%v",
			processAlive(ppid1), processAlive(child))
	}
	assertTrailingNewline(t, resp.Stdout, "default dry-run stdout")
}
```
