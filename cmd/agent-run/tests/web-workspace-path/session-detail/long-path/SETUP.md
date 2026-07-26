# Scenario

**Feature**: session with long meta.workspace (needs compact + expand)

```
# deep path in session meta (not web cwd)
makeDeepWorkspaceDir path string -> seed meta.workspace
  -> session header shows WorkspacePath for that path
```

## Preconditions

- Long absolute path string written to session `meta.workspace`.
- Directory is created so path is realistic; web process cwd need not match.

## Steps

1. Build deep path under temp; set `WorkspacePath` and default `SessionID`.
2. Leaf seeds session, starts web, runs expand script.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	deep := makeDeepWorkspaceDir(t, req.TempDir)
	req.WorkspacePath = deep
	if req.SessionID == "" {
		req.SessionID = "ws-path-session-long"
	}
	return nil
}
```
