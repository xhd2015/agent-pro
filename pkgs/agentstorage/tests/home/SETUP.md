# Scenario

**Feature**: home directory resolution and lazy directory creation

```
AGENT_RUN_HOME env -> Store.Home() resolved path
first AppendEvent/CreateSession -> mkdir sessions/<session_id>/
```

## Preconditions

- `agentstorage.NewFileStore` accepts an explicit home path but honors `AGENT_RUN_HOME` when set.
- Session subdirectories are created lazily on first write, not at store open.
- Flat layout: no `sessions/<runner>/` intermediate directory.

## Steps

1. Set `req.Operation = "home"`.
2. Leaf Setup configures env override or first-write target session id.
3. `Run` opens store and either returns resolved home or performs first write.
4. Leaf `Assert` checks home path or created directory layout.

## Context

- `req.Home` is the constructor argument; `AGENT_RUN_HOME` may point elsewhere for override tests.
- `Response.FilesWritten` lists files created under home after first write.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Operation = "home"
	if req.Runner == "" {
		req.Runner = "fake-opencode"
	}
	return nil
}
```