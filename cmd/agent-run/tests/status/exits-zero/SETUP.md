# Scenario

**Feature**: `status` subcommand exits with code 0

```
agent-run status → exit 0
```

## Preconditions

- `agent-run` binary is built (inherited from root `SETUP.md`).

## Steps

1. Run `agent-run status`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"status"}
	return nil
}
```