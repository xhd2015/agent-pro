# Scenario

**Feature**: git commit must not race with concurrent git operations from the agent

## Steps
1. Inherit harness from root SETUP.md.
2. Leaf configures fake-opencode to spawn background git loops before returning.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.TempDir == "" {
		return fmt.Errorf("commit-race subtree requires initialized TempDir from root Setup")
	}
	return nil
}
```