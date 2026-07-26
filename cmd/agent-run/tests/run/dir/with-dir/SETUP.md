# Scenario

**Feature**: `run --dir` is set (absolute, relative, or invalid)

```
agent-run run --dir <path> --agent-runner fake-codex …
  -> valid: meta.workspace = resolved abs path
  -> invalid: non-zero exit + clear stderr
```

## Steps

1. Leaves append `--dir` and prompt-specific args on top of `run --agent-runner fake-codex`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Ensure run prefix is present before leaves append --dir.
	if len(req.Args) < 1 || req.Args[0] != "run" {
		req.Args = append([]string{"run", "--agent-runner", "fake-codex"}, req.Args...)
	}
	return nil
}
```
