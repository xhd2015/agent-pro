# Scenario

**Feature**: empty Codex home yields no sessions message

```
# CodexHome exists but has no rollout JSONL files
sessions.List(codexHome, 20) -> empty slice

# table formatter prints friendly empty message
FormatListTable -> "No sessions found"
```

## Preconditions

- No files under `{CodexHome}/sessions/`.

## Steps

1. Do not write any rollout fixtures (root temp dir only).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Limit = 20
	return nil
}
```