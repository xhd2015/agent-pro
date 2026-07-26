## Expected

- `ready=false`, `state=loading` (or any non-`idle` state — assert prefers `loading`).
- Must **not** return `ready=true` or `state=idle`.
- Fixture contains update-modal markers (`Update available`, `Skip until next version` or `Press enter to continue`).
- `reason` mentions update (substring `update`) once implementer adds the signal.

## Exit Code

N/A (direct package call)

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(filepath.Join(req.TestdataDir, fixtureUpdateModal))
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(text))
	if !strings.Contains(lower, "update available") {
		t.Fatalf("fixture must contain Update available")
	}
	if !strings.Contains(lower, "skip until next version") && !strings.Contains(lower, "press enter to continue") {
		t.Fatalf("fixture must look like update modal menu")
	}
	if !strings.Contains(string(text), "›") && !strings.Contains(string(text), "\u203a") {
		t.Fatalf("fixture must contain › (the false-positive prompt marker)")
	}

	if resp.Status.Ready {
		t.Fatalf("update modal must not be ready=true (got state=%q reason=%q) — /status would be injected",
			resp.Status.State, resp.Status.Reason)
	}
	if resp.Status.State == "idle" {
		t.Fatalf("update modal must not be state=idle (got ready=%v reason=%q)",
			resp.Status.Ready, resp.Status.Reason)
	}
	// Preferred post-fix classification from requirement table.
	assertWritable(t, "update-modal-not-idle", resp.Status, false, "loading", "update")
}
```
