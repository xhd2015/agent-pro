## Expected

- `Run` returns one `FixtureResult` per `expectations.jsonl` row (currently 20 fixtures).
- Every result matches `ready`, `state`, and `reason` (when manifest specifies `reason`).
- No fixture file is missing from disk.

## Exit Code

N/A (direct package call)

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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