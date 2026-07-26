# Scenario

**Feature**: The program imports `agent/event/types` and marshals a `FileChange`

## Preconditions
- The program imports `agent/event/types` and marshals a `FileChange`.

## Steps
1. Create a FileChange with path and kind.
2. Marshal to JSON and print.

```go
import (
	"testing"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Value = types.FileChange{Path: "bar.go", Kind: "modify"}
	return nil
}
```
