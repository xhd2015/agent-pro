# Scenario

**Feature**: `pty kill-orphans --help` documents `--dry-run` and `--exe`

```
agent-run pty kill-orphans --help -> --dry-run, --exe
```

## Steps

1. Run kill-orphans help (no spawn).
2. Assert flags are documented; exit 0; trailing newline.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "" // plain CLI exec, no spawn
	req.Args = []string{"pty", "kill-orphans", "--help"}
	return nil
}
```
