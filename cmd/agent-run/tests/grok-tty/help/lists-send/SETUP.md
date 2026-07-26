# Scenario

**Feature**: top-level help documents `send` subcommand

```
agent-run --help → stdout contains send
```

## Steps

1. Run `agent-run --help`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"--help"}
	return nil
}
```