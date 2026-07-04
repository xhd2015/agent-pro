# Scenario

**Feature**: empty Grok home yields no sessions message

```
# GrokHome exists but has no summary.json files
sessions.List(grokHome, 20) -> empty slice

# table formatter prints friendly empty message
FormatListTable -> "No sessions found"
```

## Preconditions

- No files under `{GrokHome}/sessions/`.

## Steps

1. Do not write any session fixtures (root temp dir only).
2. Set `req.Limit = 20`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Limit = 20
	return nil
}
```