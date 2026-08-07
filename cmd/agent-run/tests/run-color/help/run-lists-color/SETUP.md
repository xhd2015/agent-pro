# Scenario

**Feature**: `run --help` lists `--color`

```
agent-run run --help -> documents --color
```

## Steps

1. Run `agent-run run --help`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"run", "--help"}
	return nil
}
```
