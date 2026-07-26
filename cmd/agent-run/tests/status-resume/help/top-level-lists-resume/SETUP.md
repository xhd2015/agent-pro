# Scenario

**Feature**: top-level help lists the `resume` command

```
agent-run --help -> stdout contains resume
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
