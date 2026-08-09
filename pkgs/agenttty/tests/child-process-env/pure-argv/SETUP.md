# Scenario

**Feature**: composition keeps pure agent argv (no `env` binary prefix)

```
# S9 pure argv
pureArgv=["codex", "exec"] + non-empty policy inputs
  -> ApplyChildProcessEnv (transitional) returns argv that does NOT start with "env"
  # HeadlessRun Command must remain pure agent argv
```

## Steps

1. Grouping sets Mode=apply-argv and a pure agent Argv.
2. Leaf enables a non-empty policy (color and/or home) that today would prefix `env`.
3. Assert resulting argv[0] is not `env` and matches pure command head.

## Context

- Until implementer changes ApplyChildProcessEnv (or removes env-prefix path),
  this leaf is **assertion-RED**.
- Preferred end state: Apply returns pure argv unchanged; callers use
  BuildChildProcessEnv for Set/Unset.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Mode = "apply-argv"
	if len(req.Argv) == 0 {
		req.Argv = []string{"codex", "exec"}
	}
	return nil
}
```
