# Scenario

**Feature**: gen-commit-msg with fake-opencode agent runner

## Steps
1. Inherit harness from root SETUP.md.
2. Leaf configures mock events and staged git changes.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.TempDir == "" {
		return fmt.Errorf("commit-with-fake-opencode subtree requires initialized TempDir from root Setup")
	}
	return nil
}
```