# Scenario

**Feature**: dirty agent payloads sanitize to a usable commit message

```
# accepted path: parse + sanitize yields non-empty title (and optional body)
fake-opencode dirty text -> sanitize -> stdout clean message
optional --commit -> git subject is cleaned title only
```

## Preconditions
- Parent `sanitize` helpers load fixtures and write mock agent text.
- Outcome is success (non-zero only if product fails — leaves expect success).

## Steps
1. Inherit fixture helpers.
2. Leaf picks one anti-pattern fixture and whether to `--commit`.

```go
import (
	"fmt"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.TempDir == "" {
		return fmt.Errorf("sanitize/accepted requires TempDir from root Setup")
	}
	// Default: print only; leaves may set req.Commit = true.
	req.Commit = false
	return nil
}
```
