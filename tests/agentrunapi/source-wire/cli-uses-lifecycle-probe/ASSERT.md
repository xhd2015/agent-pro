## Expected

- At least one non-test `.go` under `pkgs/agentruncli` references
  `LifecycleProbe`.
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
		t.Fatal("expected to scan pkgs/agentruncli .go files")
	}
	if !resp.LifecycleProbeFound {
		t.Fatalf("pkgs/agentruncli must reference LifecycleProbe (scanned %d files)",
			resp.ScannedFiles)
	}
}
```
