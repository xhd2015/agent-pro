# Scenario

**Feature**: Grouping node for ToPi conversion tests

## Preconditions
- Grouping node for ToPi conversion tests.

## Steps
1. Sets `Target` to `"to_pi"` as a default for all ToPi leaves.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "to_pi"
	return nil
}
```
