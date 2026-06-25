# Scenario

**Feature**: auto-unstage resolves staged paths when `--dir` is a subdirectory of the git repo root

## Preconditions
- Root harness from `agent/commit_msg/tests/SETUP.md` has initialized `req.TempDir` and built fake-opencode.

## Steps
1. Inherit harness from root SETUP.md.
2. Leaf configures a monorepo layout and runs gen-commit-msg from a nested subdirectory.

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.TempDir == "" {
		return fmt.Errorf("auto-unstage subtree requires initialized TempDir from root Setup")
	}
	return nil
}
```