# Scenario

**Feature**: live claude headless run produces a canonical message containing the expected answer

```
# prompt claude to say pong, parse stream-json, verify "pong" in the canonical output
claude -p "Reply with exactly the word: pong" --output-format stream-json --verbose -> FromClaude -> ActionMessage with "pong"
```

## Preconditions
- `claude` is available on `PATH` (or skip via `CLAUDE_SKIP_INTEGRATION=1`).

## Steps
1. Use the default `ClaudePath` (empty → `LookPath`).
2. Send the prompt `Reply with exactly the word: pong`.

## Context
- The root `runClaudeHeadless` locates the claude binary and runs it headless.
- If the binary is not found the test is skipped.
- The prompt expects the single word `pong` somewhere in the assistant text.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "Reply with exactly the word: pong"
	return nil
}
```
