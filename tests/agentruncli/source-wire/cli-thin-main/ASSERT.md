## Expected

- At least one non-test `.go` under `cmd/agent-run` was scanned.
- Import of `github.com/xhd2015/agent-pro/pkgs/agentruncli` (or `pkgs/agentruncli`) found.
- A `Handle` call site is present with that import (e.g. `agentruncli.Handle`).

## Side Effects

- None (read-only source inspection).

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.ScannedFiles == 0 {
		t.Fatal("expected to scan cmd/agent-run .go files")
	}
	if !resp.ImportFound {
		t.Fatalf("cmd/agent-run must import github.com/xhd2015/agent-pro/pkgs/agentruncli (scanned %d files)",
			resp.ScannedFiles)
	}
	if !resp.HandleRef {
		t.Fatalf("cmd/agent-run must call agentruncli.Handle (scanned %d files)",
			resp.ScannedFiles)
	}
}
```
