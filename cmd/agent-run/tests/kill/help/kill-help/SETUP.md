# Scenario

**Feature**: `kill --help` documents usage, session-id, and `--dry-run`

```
agent-run kill --help -> usage + session-id + --dry-run; exit 0; trailing newline
```

## Steps

1. Set Args to `kill --help`.
2. Run Mode handle.
3. Assert documents kill contract flags and ends with `\n`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	req.Args = []string{"kill", "--help"}
	return nil
}
```
