# Scenario

**Feature**: `resume --help` lists `--prepend-path` and `-e`/`--env`

```
agent-run resume --help -> documents --prepend-path and -e/--env
```

## Steps

1. Run `agent-run resume --help`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"resume", "--help"}
	return nil
}
```
