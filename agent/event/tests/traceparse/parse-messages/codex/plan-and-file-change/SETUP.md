# Scenario

**Feature**: traceparse leaf `parse-messages/codex/plan-and-file-change`

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
	req.RawLines = []string{
		`{"type":"item.started","item":{"id":"item_7","type":"todo_list","items":[{"text":"Inspect Jira comments","completed":false},{"text":"Write output JSON","completed":false}]}}`,
		`{"type":"item.updated","item":{"id":"item_7","type":"todo_list","items":[{"text":"Inspect Jira comments","completed":true},{"text":"Write output JSON","completed":false}]}}`,
		`{"type":"item.completed","item":{"id":"item_8","type":"file_change","changes":[{"path":"/tmp/code-commits.json","kind":"add"}],"status":"completed"}}`,
	}
	req.CreatedAt = "2026-04-28T16:57:57.816512+08:00"
	return nil
}
```
