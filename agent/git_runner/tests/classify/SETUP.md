# Scenario

**Feature**: IsTransientIndexError classifies git index write/lock messages

```
doctest -> IsTransientIndexError(ClassifyOutput, nil) -> Transient true|false
```

## Preconditions

- Mode is pure classification; no git repository required.
- Leaf sets `ClassifyOutput` and `WantTransient`.

## Steps

1. Set `req.Mode = "classify"`.
2. Leaf supplies the git error text and expected transient flag.
3. `Run` calls `IsTransientIndexError` and returns `resp.Transient`.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Mode != "" && req.Mode != "classify" {
		return fmt.Errorf("classify subtree requires Mode=classify, got %q", req.Mode)
	}
	req.Mode = "classify"
	return nil
}
```
