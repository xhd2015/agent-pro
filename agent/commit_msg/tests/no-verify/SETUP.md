# Scenario

**Feature**: `--no-verify` flag validation and git commit forwarding

```
# --no-verify alone must fail before agent
gen-commit-msg --no-verify -> flag validation error (no agent call)

# --commit --no-verify skips git hooks
staged diff -> gen-commit-msg --commit --no-verify -> fake-opencode -> git commit --no-verify
```

## Preconditions
- Root harness from `agent/commit_msg/tests/SETUP.md` has initialized `req.TempDir` and built fake-opencode.

## Steps
1. Inherit harness from root SETUP.md.
2. Leaf configures `--no-verify` / `--commit` combination and git hook state.

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.TempDir == "" {
		return fmt.Errorf("no-verify subtree requires initialized TempDir from root Setup")
	}
	return nil
}
```