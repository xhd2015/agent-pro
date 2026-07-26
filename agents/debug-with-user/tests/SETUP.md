# Scenario

**Feature**: dialog package escapes AppleScript strings and parses osascript output

```
caller text -> Escape -> AppleScript-safe literal
osascript stdout -> ParseOsascriptOutput -> button label and/or free text
```

## Preconditions

- Package `github.com/xhd2015/agent-pro/agents/debug-with-user/dialog` exports
  `Escape` and `ParseOsascriptOutput`.
- Tests call the package directly (no GUI, no subprocess).

## Steps

1. Root `Setup` sets default `req.Operation` when unset.
2. Leaf `Setup` narrows `req.Operation`, `req.Input`, and expected parse fields.
3. `Run` dispatches to `dialog.Escape` or `dialog.ParseOsascriptOutput`.
4. Leaf `Assert` checks escaped output or parsed fields.

## Context

- Escape mirrors the os-bar Swift helper: backslash and quote doubling.
- Parse accepts multi-line osascript stdout; button and text lines may appear
  together (Customize flow) or alone.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Operation == "" {
		req.Operation = "escape"
	}
	return nil
}
```
