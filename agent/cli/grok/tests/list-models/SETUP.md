# Scenario

**Feature**: grok ListModels() operation returns the available models

## Preconditions
- The grok binary is available in PATH.
- This leaf tests the `ListModels()` operation.

## Steps
1. Set `Operation` to `"list-models"` to invoke `agent.ListModels()`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Operation = OpListModels
	return nil
}
```
