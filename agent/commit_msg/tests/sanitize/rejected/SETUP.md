# Scenario

**Feature**: unusable agent payloads hard-fail after sanitize (no commit)

```
# rejected path: tool noise / empty after sanitize
fake-opencode garbage -> sanitize rejects -> non-zero error
with --commit: HEAD must not move
```

## Preconditions
- Parent `sanitize` helpers load fixtures.
- Outcome is hard failure; no LLM retry.

## Steps
1. Inherit fixture helpers.
2. Leaf configures garbage fixture and usually enables `--commit` to prove no side effect.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.TempDir == "" {
		return fmt.Errorf("sanitize/rejected requires TempDir from root Setup")
	}
	return nil
}
```
