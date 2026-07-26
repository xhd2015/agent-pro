# Scenario

**Feature**: Grouping node for wire-format event type tests

## Preconditions
- Grouping node for wire-format event type tests.

## Steps
1. Sets `Target` to `"wire"` as a default for all wire-format leaves.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "wire"
	return nil
}
```
