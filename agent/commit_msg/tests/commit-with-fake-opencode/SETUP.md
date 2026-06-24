# Scenario

**Feature**: gen-commit-msg with fake-opencode agent runner

## Steps
1. Inherit harness from root SETUP.md.
2. Leaf configures mock events and staged git changes.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	return nil
}
```