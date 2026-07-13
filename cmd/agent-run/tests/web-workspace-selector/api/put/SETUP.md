# Scenario

**Feature**: `PUT /api/agent-run/workspace` commit + validation + MRU

```
# only PUT commits selection into config + recent_workspaces
PUT /api/agent-run/workspace {"path":"..."}
  -> valid dir: 200, selected + MRU head
  -> invalid: 400, no change
```

## Preconditions

- Absolute paths under TempDir for valid/invalid cases.
- MRU cap = 12.

## Steps

1. Leaves create dirs/files and set PUT HTTPSteps.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Scenario == "" {
		req.Scenario = "put"
	}
	return nil
}
```
