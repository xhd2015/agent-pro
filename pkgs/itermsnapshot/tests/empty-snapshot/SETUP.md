# Scenario

**Feature**: empty base inventory yields empty agents

```
Snapshot with zero windows -> Capture(enrich) -> Agents empty, success
```

## Steps

1. Leaves inject empty or zero-window Snapshot under enrich (NoEnrich=false).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.NoEnrich = false
	return nil
}
```
