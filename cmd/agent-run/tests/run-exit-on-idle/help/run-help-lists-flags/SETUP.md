# Scenario

**Feature**: `agent-run run -h` lists `--exit-on-idle` and `--idle-timeout`

```
agent-run run -h -> stdout documents both flags, default 10m, sendable-prompt wording
```

## Steps

1. Set Args to `run -h`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Args = []string{"run", "-h"}
	return nil
}
```
