# Scenario

**Feature**: `CrushPath` is set to a path that does NOT exist on the filesystem

## Preconditions
- `CrushPath` is set to a path that does NOT exist on the filesystem.

## Steps
1. Override `CrushPath` to a nonexistent file.
2. Use any prompt (server will never start).

## Context
- When `CrushPath` is explicitly set and the file does not exist, the Run
  function must return a non-nil error (rather than skipping) so the caller
  can detect configuration issues.
- If `CRUSH_SKIP_INTEGRATION=1` is set in the environment the test may skip
  instead (user explicitly opted out of all integration tests).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.CrushPath = "/nonexistent/path/to/crush"
	req.Prompt = "test"
	req.HostPort = 0
	return nil
}
```
