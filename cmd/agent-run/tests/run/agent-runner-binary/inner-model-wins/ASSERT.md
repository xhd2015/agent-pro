---
label: e2e
---

## Expected

- Exit code 0.
- Captured output `ARGV_RECORD` contains `--model inner`.
- Argv record does **not** contain `--model outer`.

## Exit Code

0

```go
import (
	"os"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	assertSuccess(t, resp)
	probe, err := os.ReadFile(req.ArgvProbePath)
	if err != nil {
		t.Fatalf("read argv probe %s: %v", req.ArgvProbePath, err)
	}
	record := strings.TrimSpace(string(probe))
	if !strings.Contains(record, "--model inner") {
		t.Fatalf("ARGV_RECORD missing --model inner: %q", record)
	}
	if strings.Contains(record, "--model outer") {
		t.Fatalf("ARGV_RECORD must not contain --model outer: %q", record)
	}
}
```