## Expected

- `SaveConfig` and reload `Config()` succeed.
- Reloaded config matches the saved values exactly.
- `config.json` exists under home.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertEqual(t, "DefaultAgentRunner", resp.Config.DefaultAgentRunner, req.Config.DefaultAgentRunner)
	assertEqual(t, "DefaultModel", resp.Config.DefaultModel, req.Config.DefaultModel)
	assertEqual(t, "LastSession", resp.Config.LastSession, req.Config.LastSession)

	foundConfig := false
	for _, p := range resp.FilesWritten {
		if strings.HasSuffix(p, "config.json") {
			foundConfig = true
		}
	}
	if !foundConfig {
		t.Fatalf("expected config.json under %q, got: %v", resp.ResolvedHome, resp.FilesWritten)
	}
	want := filepath.Join(resp.ResolvedHome, "config.json")
	_ = want
}
```