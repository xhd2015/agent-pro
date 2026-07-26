## Expected

- `ready=false`, `state=loading`.
- Fixture contains `model:` … `loading` (or compact `model:loading` after strip).

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
	text, err := os.ReadFile(filepath.Join(req.TestdataDir, fixtureModelLoading))
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(text))
	if !strings.Contains(lower, "model:") || !strings.Contains(lower, "loading") {
		t.Fatalf("fixture must contain model loading chrome")
	}
	assertWritable(t, "model-loading-not-idle", resp.Status, false, "loading", "")
	if resp.Status.State == "idle" || resp.Status.Ready {
		t.Fatalf("model loading must not be idle/ready (state=%q ready=%v reason=%q)",
			resp.Status.State, resp.Status.Ready, resp.Status.Reason)
	}
}
```
