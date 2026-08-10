# Scenario

**Feature**: `takeover --help` documents usage, session-id, aliases, and options

```
agent-run takeover --help
  -> exit 0
  -> usage mentions takeover and <session-id>
  -> documents --grok, --codex, --agent-runner, --dry-run
  -> trailing newline
```

## Steps

1. Set Args to `takeover --help`.
2. Run Mode handle.
3. Assert documents takeover contract flags and ends with `\n`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	req.Args = []string{"takeover", "--help"}
	return nil
}
```
