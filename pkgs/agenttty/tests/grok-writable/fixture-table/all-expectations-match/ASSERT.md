## Expected

- `Run` returns one `FixtureResult` per `expectations.jsonl` row (currently 23 fixtures, including 3 modern open-ready frames).
- Every result matches `ready`, `state`, and `reason` (when manifest specifies `reason`).
- Optional manifest fields (`banner_detected_legacy`, `open_ready`, `screen_class`) are **ignored** by this leaf (F1 writable-only gate).
- No fixture file is missing from disk.

## Exit Code

N/A (direct package call)

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.FixtureResults) == 0 {
		t.Fatal("expected fixture table results")
	}
	manifestPath := filepath.Join(req.TestdataDir, "expectations.jsonl")
	manifest, err := loadExpectations(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.FixtureResults) != len(manifest) {
		t.Fatalf("result count %d != manifest %d", len(resp.FixtureResults), len(manifest))
	}
	for _, result := range resp.FixtureResults {
		statusMatches(t, "fixture-table", result.Expected, result.Actual)
	}
	matches, err := filepath.Glob(filepath.Join(req.TestdataDir, "grok-*.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != len(manifest) {
		t.Fatalf("grok-*.txt count %d != manifest %d", len(matches), len(manifest))
	}
	if _, err := os.Stat(filepath.Join(req.TestdataDir, "expectations.jsonl")); err != nil {
		t.Fatalf("expectations.jsonl: %v", err)
	}
}
```