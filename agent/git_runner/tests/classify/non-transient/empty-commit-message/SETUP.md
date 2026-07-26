# Scenario

**Feature**: empty commit message abort is not a transient index error

```
IsTransientIndexError("Aborting commit due to empty commit message.", nil) -> false
```

## Steps

1. Set `ClassifyOutput` to git's empty-message abort text.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClassifyOutput = "Aborting commit due to empty commit message."
	req.WantTransient = false
	return nil
}
```
