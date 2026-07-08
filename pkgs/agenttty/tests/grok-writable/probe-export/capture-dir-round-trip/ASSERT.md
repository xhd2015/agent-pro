## Expected

- Probe export exits successfully (after `-export-fixtures` is implemented).
- Export dir contains at least one `grok-*.txt` and `expectations.jsonl`.
- Every manifest row references an on-disk fixture; JSON fields `file`, `ready`, `state`, `tags` are present.

## Errors

- Until implementer adds export flags, `Run` returns non-nil error from probe (acceptable RED).

## Exit Code

N/A (go run probe subprocess)

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("probe export failed (expected RED until -export-fixtures implemented): %v\noutput:\n%s", err, resp.ProbeOutput)
	}
	if len(resp.ExportFiles) == 0 {
		t.Fatalf("expected exported grok-*.txt files; output:\n%s", resp.ProbeOutput)
	}
	if len(resp.ExportManifest) == 0 {
		t.Fatal("expected non-empty expectations.jsonl from export")
	}
	for _, exp := range resp.ExportManifest {
		if exp.File == "" || exp.State == "" {
			t.Fatalf("invalid manifest row: %+v", exp)
		}
		path := filepath.Join(req.ExportToDir, exp.File)
		if req.ExportToDir == "" {
			// Run stores export in temp; infer from first export file parent
			path = filepath.Join(filepath.Dir(resp.ExportFiles[0]), exp.File)
		}
		if _, statErr := os.Stat(path); statErr != nil {
			t.Fatalf("exported fixture missing %s: %v", exp.File, statErr)
		}
	}
}
```