## Expected

- `ready=true`, `state=idle`.
- Fixture contains main chat `›` and MCP incomplete warning.

## Exit Code

N/A (direct package call)

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(filepath.Join(req.TestdataDir, fixtureMainPromptMCP))
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(text))
	if !strings.Contains(lower, "mcp startup incomplete") {
		t.Fatalf("fixture must contain MCP startup incomplete")
	}
	if !strings.Contains(string(text), "›") && !strings.Contains(string(text), "\u203a") {
		t.Fatalf("fixture must contain main prompt ›")
	}
	assertWritable(t, "main-prompt-idle", resp.Status, true, "idle", "")
}
```
