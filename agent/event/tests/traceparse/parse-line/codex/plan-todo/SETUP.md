# Scenario

**Feature**: traceparse leaf `parse-line/codex/plan-todo`

```
trace JSONL -> adapter registry -> parsed event JSON
```

## Preconditions
- Mode and inputs are set for this leaf.

## Steps
1. Configure `Request` fields for this scenario.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RawLine = `{"type":"item.updated","item":{"id":"item_7","type":"todo_list","items":[{"text":"Inspect Jira comments","completed":true},{"text":"Write output JSON","completed":false}]}}`
	return nil
}
```
