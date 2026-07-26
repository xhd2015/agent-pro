# Scenario

**Feature**: Grouping node for ToPi conversion tests

## Preconditions
- Grouping node for ToPi conversion tests.

## Steps
1. Sets `Target` to `"to_pi"` as a default for all ToPi leaves.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "to_pi"
	return nil
}
```
