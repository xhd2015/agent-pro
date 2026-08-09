# Scenario

**Feature**: S8 — color + `-e NO_COLOR=1` drops NO_COLOR from Set; Unset has it

```
# S8 color wins over user NO_COLOR
envEntries=["NO_COLOR=1"], color=true
  -> NO_COLOR not in Set
  -> Unset contains NO_COLOR
  -> force keys still set
```

## Steps

1. Set EnvEntries to `NO_COLOR=1`.
2. ParentTERM good (`xterm`) so TERM path is non-dumb.
3. Assert NO_COLOR absent from Set; present in Unset.

## Context

- Color policy drops user `-e NO_COLOR=…` so `-u`/Unset can clear parent too.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.EnvEntries = []string{"NO_COLOR=1"}
	req.ParentTERM = "xterm"
	req.ConfigHome = ""
	req.PrependPaths = nil
	return nil
}
```
