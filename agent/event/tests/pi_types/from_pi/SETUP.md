# Scenario

**Feature**: Grouping node for FromPi conversion tests

## Preconditions
- Grouping node for FromPi conversion tests.

## Steps
1. Sets `Target` to `"from_pi"` as a default for all FromPi leaves.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Target = "from_pi"
	return nil
}
```
