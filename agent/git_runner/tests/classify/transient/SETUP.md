# Scenario

**Feature**: messages that match index lock/write races are transient

```
IsTransientIndexError(index-write or index.lock message) -> true
```

## Steps

1. Expect `WantTransient = true` for all leaves under this grouping.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Mode != "classify" {
		return fmt.Errorf("transient classify requires Mode=classify")
	}
	req.WantTransient = true
	return nil
}
```
