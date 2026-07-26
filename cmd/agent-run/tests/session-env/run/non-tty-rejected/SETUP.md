# Scenario

**Feature**: non-TTY runner + session env flags is a hard error

```
run --agent-runner fake-codex --prepend-path DIR -e FOO=bar "hi"
  -> exit ≠ 0; stderr indicates TTY-only / unsupported
```

## Steps

1. Override runner to `fake-codex` (non-TTY).
2. Pass `--prepend-path` and `-e` (either alone would suffice; both proves reject path).

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.PrependPathDir = absPath(t, filepath.Join(req.TempDir, "tools-nontty"))
	// Non-TTY hard error should happen before needing the directory or fake-codex binary.
	req.Args = []string{
		"run",
		"--agent-runner", "fake-codex",
		"--prepend-path", req.PrependPathDir,
		"-e", "FOO=bar",
		"hi",
	}
	return nil
}
```
