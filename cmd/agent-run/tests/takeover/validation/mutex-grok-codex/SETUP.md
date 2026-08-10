# Scenario

**Feature**: `--grok` and `--codex` are mutually exclusive on `takeover`

```
agent-run takeover --grok --codex <provider-session-id>
  -> exit non-zero
  -> mutually exclusive / cannot use both
```

## Steps

1. Pass both provider aliases with a positional session id (so failure is mutex, not missing id).
2. Run Mode handle.
3. Assert non-zero exit and exclusive wording.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "handle"
	req.Args = []string{
		"takeover",
		"--grok",
		"--codex",
		takeoverFixtureSessionID,
	}
	return nil
}
```
