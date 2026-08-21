# Scenario

**Feature**: env spill threshold boundary (64 vs 65 runes)

Parent for at-threshold / over-threshold leaves.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Open = true
	return nil
}
```
