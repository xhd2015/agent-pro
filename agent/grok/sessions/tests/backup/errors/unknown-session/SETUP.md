# Scenario

**Feature**: unknown session id errors before writing payload

```
# no session dir for id
Backup("019f283a-eeee-…", {GrokHome, OutDir})
  -> error "grok session not found: <id>"; OutDir has no payload
```

## Preconditions

- Grok home has empty sessions tree (or no matching id).
- Explicit OutDir path for “no payload written” assert.

## Steps

1. Set unknown `SessionID`.
2. Set `OutDir` to a fresh path (not created).

```go
import (
	"path/filepath"
	"testing"
)

const unknownBackupSessionID = "019f283a-eeee-7eee-eeee-eeeeeeeeee99"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = unknownBackupSessionID
	req.OutDir = filepath.Join(req.TempDir, "backup-unknown-out")
	// Ensure inactive + empty live injectables (root default).
	writeActiveSessions(t, req.GrokHome /* none */)
	return nil
}
```
