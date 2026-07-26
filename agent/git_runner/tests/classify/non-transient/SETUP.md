# Scenario

**Feature**: permanent commit failures must not be treated as transient

```
IsTransientIndexError(empty message | hook failure) -> false
```

## Steps

1. Expect `WantTransient = false` for all leaves under this grouping.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Mode != "classify" {
		return fmt.Errorf("non-transient classify requires Mode=classify")
	}
	req.WantTransient = false
	return nil
}
```
