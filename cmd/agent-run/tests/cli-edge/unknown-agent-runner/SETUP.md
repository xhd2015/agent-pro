# Scenario

**Feature**: unknown `--agent-runner` exits with code 1 and stderr hint

```
# L2 in-process: Handle run --agent-runner totally-bogus-runner hi
# → early validateRunner error (no agent process)
```

## Preconditions

- No binary build: `req.Mode = "handle"` uses `pkgs/agentruncli.Handle`.
- Validation fails before store open / runner spawn.

## Steps

1. Set Mode handle and args for bogus runner.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Mode = "handle"
	req.Args = []string{
		"run",
		"--agent-runner", "totally-bogus-runner",
		"hi",
	}
	return nil
}
```