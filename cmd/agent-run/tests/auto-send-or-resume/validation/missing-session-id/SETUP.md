# Scenario

**Feature**: `--auto-send-or-resume` without `--session` / `--session-id` is an error

```
agent-run run --auto-send-or-resume "hello"
  -> exit 1; stderr requires --session-id
```

## Steps

1. Run auto path with prompt but no session-id flag.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"run",
		"--auto-send-or-resume",
		"hello without session id",
	}
	return nil
}
```
