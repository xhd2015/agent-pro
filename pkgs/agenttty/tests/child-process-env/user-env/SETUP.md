# Scenario

**Feature**: user `-e` KEY=VALUE entries appear in Set

```
# user env
envEntries=["FOO=bar"]
  -> FOO=bar in Set; Unset empty (color false)
```

## Steps

1. Grouping sets Mode=build, Color=false.
2. Leaf supplies concrete EnvEntries.

## Context

- Last-wins for duplicate keys is production merge behavior; single assignment here.
- Empty/whitespace entries are skipped by production (not asserted as a leaf).

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
	req.PrependPaths = nil
	req.ParentTERM = ""
	if req.RunnerID == "" {
		req.RunnerID = "grok-tty"
	}
	return nil
}
```
