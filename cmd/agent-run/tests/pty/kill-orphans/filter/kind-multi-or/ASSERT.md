## Expected

- Exit code 0.
- Stdout lists both the TestGenerated-marked serve PID and the workdir-only serve PID.
- Dry-run leaves both alive; trailing `\n`.

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
	if resp.SpawnPIDs == nil {
		t.Fatalf("expected SpawnPIDs")
	}
	tg := resp.SpawnPIDs["tg"]
	tmp := resp.SpawnPIDs["tmp"]
	if tg <= 0 || tmp <= 0 {
		t.Fatalf("expected tg and tmp PIDs; got %#v", resp.SpawnPIDs)
	}
	out := resp.Stdout
	if !strings.Contains(out, strconv.Itoa(tg)) {
		t.Fatalf("multi-kind dry-run must list TestGenerated serve pid %d; stdout:\n%s", tg, out)
	}
	if !strings.Contains(out, strconv.Itoa(tmp)) {
		t.Fatalf("multi-kind dry-run must list workdir serve pid %d; stdout:\n%s", tmp, out)
	}
	if !processAlive(tg) || !processAlive(tmp) {
		t.Fatalf("dry-run must not kill serves")
	}
	assertTrailingNewline(t, resp.Stdout, "kind-multi-or dry-run stdout")
}
```
