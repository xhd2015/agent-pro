# Scenario

**Feature**: gen-commit-msg with fake-opencode agent runner

## Steps
1. Inherit harness from root SETUP.md.
2. Leaf configures mock events and staged git changes.

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.TempDir == "" {
		return fmt.Errorf("commit-with-fake-opencode subtree requires initialized TempDir from root Setup")
	}
	return nil
}
```