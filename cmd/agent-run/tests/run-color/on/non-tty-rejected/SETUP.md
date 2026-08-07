# Scenario

**Feature**: non-TTY runner + `--color` is a hard error

```
run --agent-runner fake-codex --color "hi"
  -> exit ≠ 0; stderr indicates TTY-only / unsupported
```

## Steps

1. Override args for non-TTY `fake-codex` with `--color` (no env-logger needed;
   hard error should occur before spawning a real agent binary).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Replace grouping default with explicit non-TTY shape (still Color ON).
	req.Args = []string{
		"run",
		"--agent-runner", "fake-codex",
		"--color",
		"hi",
	}
	return nil
}
```
