# Scenario

**Feature**: unknown session id returns not-found error

```
empty/unrelated grok home + existing target
-> RelocateCWD(unknown-id, target, {GrokHome})
-> error mentioning session id
```

## Preconditions

- No session directory exists for the requested id.
- Target directory exists (so failure is about the session, not the target).

## Steps

1. Create target workspace directory.
2. Do **not** seed any session for `req.SessionID`.
3. Set `req.SessionID` to a known missing UUID.

```go
import (
	"path/filepath"
	"testing"
)

const missingSessionID = "019f283a-bbbb-7bbb-bbbb-bbbbbbbbbb02"

func Setup(t *testing.T, req *Request) error {
	target := filepath.Join(req.TempDir, "ws-target")
	mustMkdir(t, target)
	req.TargetDir = absPath(t, target)
	req.SessionID = missingSessionID
	return nil
}
```
