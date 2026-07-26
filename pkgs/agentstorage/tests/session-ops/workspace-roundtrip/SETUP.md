# Scenario

**Feature**: session meta persists `workspace` through create and get

```
CreateSession(meta.workspace=/path/to/project) -> GetSession -> same workspace field
```

## Preconditions

- `SessionMeta` includes optional `workspace` JSON field on `meta.json`.
- Session id is unique under the test home.

## Steps

1. Set `req.Action = "workspace_roundtrip"`.
2. Set `req.SessionID` and `req.Workspace` to a distinct absolute-style path string.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "workspace_roundtrip"
	req.SessionID = "sess_workspace"
	req.Workspace = filepath.Join(req.TempDir, "project-workspace")
	return nil
}
```
