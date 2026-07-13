# Scenario

**Feature**: target path that is a regular file is rejected

```
valid session + target is a file (not a directory)
-> RelocateCWD(id, file-path, {GrokHome})
-> error; session dir remains at old location
```

## Preconditions

- Session fixture exists and is inactive.
- Target path exists as a regular file.

## Steps

1. Create old workspace and seed session.
2. Write a regular file at the intended target path.
3. Set `req.TargetDir` to that file path.

```go
import (
	"path/filepath"
	"testing"
)

const targetNotDirSessionID = "019f283a-dddd-7ddd-dddd-dddddddddd04"

func Setup(t *testing.T, req *Request) error {
	oldWS := filepath.Join(req.TempDir, "ws-old")
	mustMkdir(t, oldWS)
	req.OldCWD = absPath(t, oldWS)
	req.SessionID = targetNotDirSessionID

	targetFile := filepath.Join(req.TempDir, "ws-is-a-file")
	mustWriteFile(t, targetFile, "not a directory\n")
	req.TargetDir = absPath(t, targetFile)

	req.SessionDir = writeRelocateSession(t, req.GrokHome, req.SessionID, req.OldCWD, relocateSessionOpts{
		Title:              "target not dir",
		WritePromptContext: true,
	})
	return nil
}
```
