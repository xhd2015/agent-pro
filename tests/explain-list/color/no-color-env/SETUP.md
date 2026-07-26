# Scenario

**Feature**: NO_COLOR without --color disables ANSI

```
# NO_COLOR=1; explain list (no --color, non-TTY)
-> plain output, no ESC sequences
```

## Preconditions

- One short session.
- EnvExtra includes `NO_COLOR=1`.

## Steps

1. Seed fixture; set NO_COLOR; plain list args.
2. Assert no `\x1b` in stdout.

## Context

- Honors NO_COLOR when `--color` is absent.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"list"}
	req.EnvExtra = []string{"NO_COLOR=1"}
	req.Sessions = []SessionSeed{colorFixtureSession()}
	return nil
}
```
