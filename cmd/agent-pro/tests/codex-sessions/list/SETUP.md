# Scenario

**Feature**: list Codex rollout sessions from synthetic home

```
# walk CodexHome/sessions/**/*.jsonl and parse session_meta
sessions.List(codexHome, limit) -> []Session sorted newest first

# format as table or JSON for CLI output
[]Session -> FormatListTable / FormatListJSON -> terminal text
```

## Preconditions

- This branch tests the `list` operation only.
- Rollout fixtures are created by descendant leaf Setup functions.

## Steps

1. Set `req.Operation = "list"`.
2. Leaf Setup writes rollout JSONL files as needed for the scenario.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "list"
	return nil
}
```