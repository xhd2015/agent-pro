# Scenario

**Feature**: prependPaths produce PATH assignment joined ahead of process PATH

```
# PATH prepend
prependPaths=[/opt/a, /opt/b]
  -> Set contains PATH=/opt/a:/opt/b:… (or OS separator)
```

## Steps

1. Grouping sets Mode=build, Color=false, no config home.
2. Leaf supplies concrete prependPaths.

## Context

- Production joins with `os.PathListSeparator` and appends current process PATH
  when non-empty. Tests assert prefix membership, not full equality (parallel-safe;
  no t.Setenv of PATH).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Mode = "build"
	req.Color = false
	req.ConfigHome = ""
	req.EnvEntries = nil
	req.ParentTERM = ""
	if req.RunnerID == "" {
		req.RunnerID = "grok-tty"
	}
	return nil
}
```
