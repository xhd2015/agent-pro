# Scenario

**Feature**: pure child env policy for HeadlessRun without `env(1)` argv prefix

```
# policy builder (L2)
runnerID, configHome, prependPaths, envEntries, color, parentTERM
  -> BuildChildProcessEnv
  -> ChildEnvSpec{Set, Unset}

# composition (HeadlessRun wiring — not process-spawned here)
pureArgv + Set + Unset
  -> Command / CommandEnv / CommandUnset
  # never: env -u … KEY=val -- codex …
```

## Preconditions

1. Package `github.com/xhd2015/agent-pro/pkgs/agenttty` is importable.
2. Implementer provides (RED until present):
   - `type ChildEnvSpec struct { Set, Unset []string }`
   - `BuildChildProcessEnv(runnerID, configHome string, prependPaths, envEntries []string, color bool, parentTERM string) ChildEnvSpec`
3. Locked policy matches old `ApplyChildProcessEnv` **without** the `env` binary:
   - user `-e` last-wins; color drops `-e NO_COLOR=…` from Set
   - PATH prepend when prependPaths non-empty
   - CODEX_HOME / GROK_HOME from configHome + runnerID
   - color: Unset `NO_COLOR`; Set `FORCE_COLOR=1`, `CLICOLOR=1`, `CLICOLOR_FORCE=1`;
     TERM rewrite when effective TERM empty/`dumb` → `xterm-256color`
4. Tests inject `ParentTERM` — **no** `t.Setenv`, `os.Setenv`, `t.Chdir`, or
   process-global environ mutation (parallel-safe).
5. PATH still joins against process PATH (same as production today); leaves assert
   prefix membership, not full PATH equality.

## Steps

1. Grouping nodes narrow Mode / Color / RunnerID / ConfigHome as needed.
2. Leaves set concrete `EnvEntries`, `PrependPaths`, `ParentTERM`, `Argv`.
3. Root `Run` calls `BuildChildProcessEnv` or `ApplyChildProcessEnv`.
4. Assert inspects Set/Unset membership or argv purity via helpers below.

## Context

- **Env format**: `KEY=value` in Set; bare keys in Unset (case-sensitive Unix).
- **TERM default**: exactly `xterm-256color` when rewrite applies.
- **Good TERM**: present, non-empty, not `dumb` (e.g. `xterm`, `xterm-256color`).
- **Out of scope**: full HeadlessRun process spawn; P4 replace smoke.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	if req.Mode == "" {
		req.Mode = "build"
	}
	return nil
}

// setGet returns the last value for key in a KEY=value Set slice.
func setGet(set []string, key string) (string, bool) {
	prefix := key + "="
	found := false
	val := ""
	for _, e := range set {
		if strings.HasPrefix(e, prefix) {
			found = true
			val = e[len(prefix):]
		}
	}
	return val, found
}

func setHas(set []string, key string) bool {
	_, ok := setGet(set, key)
	return ok
}

func unsetHas(unset []string, key string) bool {
	for _, k := range unset {
		if k == key {
			return true
		}
	}
	return false
}

func assertBuildOK(t *testing.T, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("BuildChildProcessEnv harness error: %v", err)
	}
	if resp == nil {
		t.Fatal("resp is nil")
	}
}

func assertSetExact(t *testing.T, set []string, key, want string) {
	t.Helper()
	got, ok := setGet(set, key)
	if !ok {
		t.Fatalf("Set missing %s; Set=%#v", key, set)
	}
	if got != want {
		t.Fatalf("Set %s=%q, want %q; Set=%#v", key, got, want, set)
	}
}

func assertSetAbsent(t *testing.T, set []string, key string) {
	t.Helper()
	if setHas(set, key) {
		v, _ := setGet(set, key)
		t.Fatalf("Set must not contain %s (got %s=%q); Set=%#v", key, key, v, set)
	}
}

func assertUnsetHas(t *testing.T, unset []string, key string) {
	t.Helper()
	if !unsetHas(unset, key) {
		t.Fatalf("Unset missing %q; Unset=%#v", key, unset)
	}
}

func assertUnsetEmpty(t *testing.T, unset []string) {
	t.Helper()
	if len(unset) != 0 {
		t.Fatalf("Unset want empty, got %#v", unset)
	}
}

func assertSetEmpty(t *testing.T, set []string) {
	t.Helper()
	if len(set) != 0 {
		t.Fatalf("Set want empty, got %#v", set)
	}
}
```
