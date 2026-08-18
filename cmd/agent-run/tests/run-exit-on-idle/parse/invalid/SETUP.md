# Scenario

**Feature**: invalid `--idle-timeout` fails parse (non-zero, `Error:`)

```
ParseRunIdle -> error
stderr: Error: …
exit 1
```

## Steps

1. Grouping documents the invalid family (negative vs unparseable).
2. Leaves set the raw timeout string and whether `--exit-on-idle` is on.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = opParse
	return nil
}
```
