# Scenario

**Feature**: listing a store root that was never created reports no snapshots

```
list ~/.agent-pro/usages    # before the first collect ever ran
List(0) -> no records, no error
```

## Preconditions

- `req.Root` is a path under a fresh temp directory that nothing created.
- A first-time user asks for history before any collect ran; that must read as
  "nothing yet", not as a failure.

## Steps

1. Keep the root as configured; nothing creates it.

```go
import (
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.ListLasts = []int{0}
return nil
}
```
