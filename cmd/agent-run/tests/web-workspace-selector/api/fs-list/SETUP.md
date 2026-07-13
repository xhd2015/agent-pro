# Scenario

**Feature**: `GET /api/agent-run/fs/list` directory listing with typed entries

```
GET /api/agent-run/fs/list?path=<dir>
  -> entries: dirs + files; type dir|file; parent path
```

## Preconditions

- Fixture directory tree under TempDir (subdir + note.txt).

## Steps

1. Leaves build fixture and probe list.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Scenario == "" {
		req.Scenario = "fs-list"
	}
	return nil
}
```
