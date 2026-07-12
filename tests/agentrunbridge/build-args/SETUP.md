# Scenario

**Feature**: pure `BuildArgs(RunOpts)` argv mapping (no exec)

```
RunOpts flags -> BuildArgs -> []string starting with "run"
# equals-form: --session-id= --agent-runner= --dir=
```

## Preconditions

- `BuildArgs` is pure and does not call LookPath or exec.
- Binary name is **not** included in the returned slice.
- Empty `AgentRunner` omits `--agent-runner` entirely.

## Steps

1. Set `req.Mode = "build_args"`.
2. Leaf fills `RunOpts` fields on `Request`.
3. `Run` returns `Response.Args` from `BuildArgs`.

## Context

- Assertions compare exact argv order where SeaTalk parity matters
  (session → runner → auto-send → new-terminal → optional dir/nosubmit → open → `--` → prompt).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "build_args"
	return nil
}
```
