# Scenario

**Feature**: `tty-watch run --session-id <id>` reserves a user-chosen registry id

```
# custom id before command argv
harness -> tty-watch run --session-id <id> <cmd> -> registry/<id>.json
# duplicate live -> error; stale -> prune + reuse; invalid id -> validation error
tty-watch list -> all *.json ids (custom + session-N)
```

## Preconditions

- Isolated `TTY_WATCH_HOME` from root Setup.
- `--session-id` must appear before `<command>` on the `run` subcommand only.

## Steps

1. Leaf sets `req.Phase` and `req.CustomSessionID` (default `test-with-grok` when unset).
2. Harness builds argv as `run --session-id <id> <command>...` when `CustomSessionID` is set.
3. Detached flows use Ctrl-] (`\x1d`); duplicate/stale/invalid flows use direct CLI invocation where noted.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.CustomSessionID == "" {
		req.CustomSessionID = "test-with-grok"
	}
	return nil
}
```