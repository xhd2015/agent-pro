# Scenario

**Feature**: empty policy inputs produce empty Set/Unset

```
# no color, home, -e, or prepend
BuildChildProcessEnv(…, color=false, parentTERM="")
  -> ChildEnvSpec{Set:[], Unset:[]}
```

## Steps

1. Grouping marks empty-policy surface (Mode=build, Color=false, no home/paths/entries).
2. Leaf keeps all optional inputs empty.

## Context

- Baseline for identity: pure builder must not invent color/TERM/home keys.

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
	req.EnvEntries = nil
	req.ParentTERM = ""
	if req.RunnerID == "" {
		req.RunnerID = "grok-tty"
	}
	return nil
}
```
