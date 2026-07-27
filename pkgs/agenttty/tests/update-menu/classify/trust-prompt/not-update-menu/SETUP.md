# Scenario

**Bug**: trust modal shares "Press enter to continue" with update menu footer and
was misclassified as `IsBlockingUpdateMenu` → `sendable: no (codex update available)`.

```
06-trust-prompt.snapshot.txt
  -> IsBlockingUpdateMenu=false
  -> CheckWritable reason mentions trust (not update available)
```

## Steps

1. `FixtureFile=06-trust-prompt.snapshot.txt`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.FixtureFile = "06-trust-prompt.snapshot.txt"
	return nil
}
```
