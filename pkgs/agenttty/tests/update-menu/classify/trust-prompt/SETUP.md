# Scenario

**Factor**: directory trust prompt vs Update available menu

```
trust modal ("Do you trust…", "Yes, continue", "Press enter to continue")
  != Update available menu
```

## Steps

1. Leaf loads trust fixture and classifies.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = req
	return nil
}
```
