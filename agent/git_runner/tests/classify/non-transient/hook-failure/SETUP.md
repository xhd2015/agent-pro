# Scenario

**Feature**: pre-commit / husky hook failure is not a transient index error

```
IsTransientIndexError("husky - pre-commit hook exited with code 1", nil) -> false
```

## Steps

1. Set `ClassifyOutput` to a typical husky pre-commit failure line.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClassifyOutput = "husky - pre-commit hook exited with code 1"
	req.WantTransient = false
	return nil
}
```
