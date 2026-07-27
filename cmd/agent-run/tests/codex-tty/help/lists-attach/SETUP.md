# Scenario

**Feature**: top-level help documents `attach` subcommand

```
agent-run --help → stdout contains attach
```

## Steps

1. Run `agent-run --help`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"--help"}
	return nil
}
```