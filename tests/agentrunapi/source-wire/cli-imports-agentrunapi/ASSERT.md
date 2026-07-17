## Expected

- At least one non-test `.go` under `cmd/agent-run` imports
  `github.com/xhd2015/agent-pro/pkgs/agentrunapi`.
- ScannedFiles > 0.

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
	if resp.ScannedFiles == 0 {
		t.Fatal("expected to scan cmd/agent-run .go files")
	}
	if !resp.ImportFound {
		t.Fatalf("cmd/agent-run must import github.com/xhd2015/agent-pro/pkgs/agentrunapi (scanned %d files)",
			resp.ScannedFiles)
	}
}
```
