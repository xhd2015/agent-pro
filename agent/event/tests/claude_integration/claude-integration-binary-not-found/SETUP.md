# Scenario

**Feature**: an explicitly-configured missing ClaudePath surfaces a non-nil error

```
# ClaudePath set + os.Stat fails -> runClaudeHeadless returns error (not skip)
Run{Target:"claude_headless", ClaudePath:"/nonexistent/claude"} -> error referencing missing binary
```

## Preconditions
- `ClaudePath` is set to a path that does NOT exist on the filesystem.

## Steps
1. Override `ClaudePath` to a nonexistent file.
2. Use any prompt (claude will never run).

## Context
- When `ClaudePath` is explicitly set and the file does not exist, `runClaudeHeadless` must return a non-nil error (rather than skipping) so the caller can detect configuration issues.
- If `CLAUDE_SKIP_INTEGRATION=1` is set the test may skip instead.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClaudePath = "/nonexistent/claude"
	req.Prompt = "test"
	return nil
}
```
