# Scenario

**Feature**: `Run(RunOpts)` with fake LookPath / RunCommand / RunOutput

```
LookPath(binary) -> RunCommand|RunOutput(launch) -> [RunOutput(tty status)*]
  -> RunResult | error
```

## Preconditions

- All leaves inject hooks; no real PATH binary.
- Default scripted LookPath success path is `/fake/agent-run`.
- Wait-ready leaves set short `ReadyTimeout` / `ReadyPollInterval` and script status polls.

## Steps

1. Set `req.Mode = "run"`.
2. Leaf configures opts + hook scripts.
3. Assert errors, call counts, argv, or stdout.

## Context

- API errors land in `resp.ErrString` (harness error nil).
- `LaunchCalls` counts primary run invocations (RunCommand or capturing RunOutput).
- `StatusPollCalls` counts `tty status` polls only.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "run"
	return nil
}
```
