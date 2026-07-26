# Scenario

**Feature**: Grouping node for roundtrip conversion tests

## Preconditions
- Grouping node for roundtrip conversion tests.

## Steps
1. Sets `Target` to `"roundtrip"` as a default for all roundtrip leaves.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "roundtrip"
	return nil
}
```
