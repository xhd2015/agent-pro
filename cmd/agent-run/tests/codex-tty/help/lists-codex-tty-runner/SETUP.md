# Scenario

**Feature**: runner help lists `codex-tty` as a supported backend

```
agent-run run --help → stdout contains codex-tty
```

## Steps

1. Run `agent-run run --help`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"run", "--help"}
	return nil
}
```
