# Scenario

**Feature**: PUT path to a regular file returns 400 (A3)

```
PUT /workspace {"path":"/…/note.txt"}  # exists but is a file
  -> 400
  -> baseline selected_workspace unchanged
```

## Preconditions

- Baseline selected from parent.
- Existing regular file under TempDir.

## Steps

1. Write file; start web; PUT file path.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "file-path-400"
	filePath := filepath.Join(req.TempDir, "workspaces", "not-a-dir.txt")
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filePath, []byte("not a dir\n"), 0644); err != nil {
		return err
	}
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}
	req.HTTPSteps = []HTTPStep{
		{Name: "put", Method: "PUT", Path: "/api/agent-run/workspace", Body: putWorkspaceBody(filePath)},
	}
	return nil
}
```
