# Scenario

**Feature**: home with deeply nested workspace cwd (needs compact label)

```
# deep process cwd -> long status.workspace
makeDeepWorkspaceDir -> agent-run web cmd.Dir=deep
  -> GET /api/agent-run/status.workspace is long
  -> collapsed label uses shortWorkspaceLabel (…/last/two)
```

## Preconditions

- `req.WebWorkingDir` is a deep absolute path under `t.TempDir()`.
- Path has more than 2 segments so collapsed form starts with `…/`.

## Steps

1. Create deep workspace directory; set `WebWorkingDir` and `WorkspacePath`.
2. Leaf starts web and runs interaction-specific Playwright script.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	deep := makeDeepWorkspaceDir(t, req.TempDir)
	req.WebWorkingDir = deep
	req.WorkspacePath = deep
	return nil
}
```
