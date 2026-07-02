# Scenario

**Feature**: matching Codex rollout transcript is present

```
fake Codex TUI scrollback includes codex resume <uuid>
  -> matching rollout JSONL exists under CODEX_HOME
  -> agent-run tails the rollout file for assistant records
```

## Preconditions

- The rollout file name contains the same UUID printed in the resume line.
- The fake TUI does not print the assistant answer itself, so stdout must come from JSONL.

## Steps

1. Create the rollout directory before `agent-run` starts.
2. Configure the fake TUI to print the resume line and stay alive briefly.
3. Leaf setup seeds or schedules transcript records.

## Context

- Sibling leaves split by Codex JSONL record kind and stream timing behavior.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	ensureCodexTranscriptPath(t, req)
	req.CodexTTYCommand = fakeTUICodexResumeThenSleep(1)
	req.StreamProbeSubstring = ""
	return nil
}
```
