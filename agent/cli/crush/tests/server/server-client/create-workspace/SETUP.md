# Scenario

**Feature**: createWorkspace POST creates a workspace for the current working directory

## Preconditions
- Server is running (via `ensureServer`).
- A workspace can be created via `POST /v1/workspaces` with the current working directory.

## Steps
1. Set `ServerOperation` to `"create-workspace"`.
2. Root `Run` will:
   a. Ensure server is running.
   b. Call `createWorkspace` with a temp directory.
   c. Return the workspace ID.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ServerOperation = "create-workspace"
	return nil
}
```
