# Scenario

**Feature**: ListModels() returns nil because claude has no model-listing command

```
# claude exposes no --list-models; ListModels is unsupported
list-models -> ClaudeAgent.ListModels()
ClaudeAgent -> (no claude invocation) -> nil models, nil error
```

## Preconditions
- The `claude` binary is available in PATH (resolved by Run, skipped if absent).
- This leaf tests the `ListModels()` operation.

## Steps
1. Set `Operation` to `"list-models"` to invoke `agent.ListModels()`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = OpListModels
	return nil
}
```
