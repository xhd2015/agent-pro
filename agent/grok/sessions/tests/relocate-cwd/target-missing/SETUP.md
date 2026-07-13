# Scenario

**Feature**: missing target directory is rejected

```
valid session + target path does not exist
-> RelocateCWD(id, missing-path, {GrokHome})
-> error; session dir remains at old location
```

## Preconditions

- Session fixture exists and is inactive.
- `req.TargetDir` points at a path that does not exist.

## Steps

1. Create old workspace and seed session.
2. Point `TargetDir` at a non-existent path under temp.
3. Do not create that path.

```go
import (
	"path/filepath"
	"testing"
)

const targetMissingSessionID = "019f283a-cccc-7ccc-cccc-cccccccccc03"

func Setup(t *testing.T, req *Request) error {
	oldWS := filepath.Join(req.TempDir, "ws-old")
	mustMkdir(t, oldWS)
	req.OldCWD = absPath(t, oldWS)
	req.SessionID = targetMissingSessionID
	req.TargetDir = absPath(t, filepath.Join(req.TempDir, "ws-does-not-exist"))

	req.SessionDir = writeRelocateSession(t, req.GrokHome, req.SessionID, req.OldCWD, relocateSessionOpts{
		Title:              "target missing",
		WritePromptContext: true,
		UpdatesBody:        `{"type":"init"}` + "\n",
	})
	return nil
}
```
