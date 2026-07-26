# Scenario

**Feature**: `pty stats` prints a useful human-readable summary and exits 0

```
agent-run pty stats -> PTY limit and/or serve-related summary; trailing \n
```

## Preconditions

- Optional probes (sysctl, lsof) may be partial; exit 0 when partial stats print.

## Steps

1. Run `agent-run pty stats`.
2. Assert exit 0, keywords for limit/PTY/serve, trailing newline.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"pty", "stats"}
	return nil
}
```
