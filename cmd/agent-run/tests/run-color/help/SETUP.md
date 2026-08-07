# Scenario

**Feature**: help text documents `run --color`

```
agent-run run --help -> documents --color
```

## Steps

1. Leaf sets `req.Args` to `run --help`.
2. Assert stdout mentions `--color` and ends with trailing `\n`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if len(req.Args) == 0 {
		req.Args = []string{"run", "--help"}
	}
	return nil
}
```
