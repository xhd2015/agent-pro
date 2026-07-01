# Scenario

**Feature**: Grouping node for roundtrip conversion tests

## Preconditions
- Grouping node for roundtrip conversion tests.

## Steps
1. Sets `Target` to `"roundtrip"` as a default for all roundtrip leaves.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Target = "roundtrip"
	return nil
}
```
