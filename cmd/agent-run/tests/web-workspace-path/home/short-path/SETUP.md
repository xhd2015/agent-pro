# Scenario

**Feature**: home with short workspace path (≤2 segments after split)

```
# short cwd — compact label equals readable path form
makeShortWorkspaceDir (…/ws/ab) -> web cmd.Dir=short
  -> collapsed label is still fully readable (no …/ needed for 2 parts)
```

## Preconditions

- Path has at most 2 non-empty segments in the shortWorkspaceLabel sense
  (helper creates `…/ws/ab` under temp — absolute path has more OS segments,
  but leaf asserts the displayed short form is non-empty and matches
  `shortWorkspaceLabel` of the server cwd).

## Steps

1. Create short workspace dir; set `WebWorkingDir` / `WorkspacePath`.
2. Leaf opens home and asserts readable collapsed label.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	short := makeShortWorkspaceDir(t, req.TempDir)
	req.WebWorkingDir = short
	req.WorkspacePath = short
	return nil
}
```
