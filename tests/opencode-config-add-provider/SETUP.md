# Scenario

**Feature**: agent-pro opencode config add-provider writes a v1 provider entry

```
# build agent-pro once, run with isolated HOME so global config lands in temp
caller args -> agent-pro opencode config add-provider --id ... -> opencode config file
# global target: $HOME/.config/opencode/opencode.json  |  --dir <d>: <d>/.opencode/opencode.json
doctest <- provider[id] = { npm, name, options.baseURL, models }
```

## Preconditions

- `go` is available in PATH (to build `./cmd/agent-pro`). Tests skip otherwise.
- `cmd/agent-pro/main.go` exists with the `opencode config add-provider` leaf
  wired into `handleOpenCodeConfig` (added by the implementer).
- The opencode config layer `agent/opencode/config/config.go` (`ReadDir`,
  `Write`) is available for leaf assertions.

## Steps

1. Build the `agent-pro` binary once (cached across the process) into a temp
   path via `buildAgentPro`.
2. Create an isolated `HOME` under `t.TempDir()` so the default global config
   target `$HOME/.config/opencode/opencode.json` lands in temp and never
   touches the developer's real `~/.config/opencode`.
3. Set `req.Bin` and `req.Home`; leaves fill in `req.Args` (and any
   `PreConfig` / `WorkDir`).

## Context

- `HOME` is the sole home-directory source (`os.UserHomeDir()` reads it), so
  overriding it fully isolates the global target.
- Each leaf gets its own fresh `HOME` (per-test `t.TempDir()`), so tests are
  independent and re-runnable.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("skipping: go not found in PATH")
	}
	bin, err := buildAgentPro(t)
	if err != nil {
		return err
	}
	req.Bin = bin
	req.Home = filepath.Join(t.TempDir(), "fake-home")
	if err := os.MkdirAll(req.Home, 0o755); err != nil {
		return err
	}
	return nil
}
```
