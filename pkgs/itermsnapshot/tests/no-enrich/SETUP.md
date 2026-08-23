# Scenario

**Feature**: NoEnrich skips all agent attach

```
CaptureOpts.NoEnrich=true -> Capture skips ResolveFromPID -> Agents empty
```

## Steps

1. Set `NoEnrich=true` for the subtree.
2. Leaves still supply busy Snapshot + resolve that would hit, to prove skip.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.NoEnrich = true
	return nil
}
```
