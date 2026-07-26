# Scenario

**Feature**: classic index.lock File exists contention is transient

```
IsTransientIndexError("Unable to create '.../index.lock': File exists", nil) -> true
```

## Preconditions

- Message mentions `index.lock` (path may vary).

## Steps

1. Set `ClassifyOutput` to a representative git lock error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClassifyOutput = "fatal: Unable to create '/repo/.git/index.lock': File exists."
	req.WantTransient = true
	return nil
}
```
