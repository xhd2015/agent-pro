## Expected

- Prompt contains the schema example and the result file path.
- Result file is absolute and not under WorkspaceDir.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	assertNoAPIError(t, resp)
	assertContains(t, resp.LaunchPrompt, req.SchemaExample)
	if resp.LaunchResultFile == "" {
		t.Fatal("expected ResultFile on launch opts")
	}
	assertContains(t, resp.LaunchPrompt, resp.LaunchResultFile)
	if !filepath.IsAbs(resp.LaunchResultFile) {
		t.Fatalf("ResultFile should be absolute, got %q", resp.LaunchResultFile)
	}
	ws := resp.LaunchWorkspace
	if ws == "" {
		ws = req.WorkspaceDir
	}
	if ws != "" && (resp.LaunchResultFile == ws || strings.HasPrefix(resp.LaunchResultFile, ws+string(filepath.Separator))) {
		t.Fatalf("ResultFile %q must not be under WorkspaceDir %q", resp.LaunchResultFile, ws)
	}
}
```
