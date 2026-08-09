# Scenario

**Feature**: S9 — Apply path does not return argv starting with `env`

```
# S9
ApplyChildProcessEnv(["codex","exec"], runner, configHome, nil, nil, color=true)
  -> out[0] != "env"
  -> out starts with pure agent argv ["codex","exec"] (prefix equal)
```

## Steps

1. Color true + config home so old Apply would emit `env -u NO_COLOR …`.
2. Assert no `env` binary prefix; pure argv preserved as command head.

## Context

- RED while ApplyChildProcessEnv still prefixes `env`.
- Implementer may delete Apply and switch this leaf to assert pure Argv + Build
  composition; until then assert on Apply return value.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Mode = "apply-argv"
	req.Argv = []string{"codex", "exec"}
	req.RunnerID = "codex-tty"
	req.ConfigHome = "/tmp/agent-pro-child-env-pure-argv"
	req.Color = true
	req.PrependPaths = nil
	req.EnvEntries = nil
	req.ParentTERM = "xterm"
	return nil
}
```
