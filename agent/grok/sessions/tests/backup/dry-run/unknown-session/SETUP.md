# Scenario

**Feature**: dry-run unknown session id errors before planning writes

```
# DryRun=true; SessionID not present under grok home
Backup(unknown-id, {DryRun, OutDir})
  -> error "grok session not found: <id>"; OutDir not created
```

## Preconditions

- Session id does not exist under fixture grok home.
- DryRun true; OutDir set for no-write assert.

## Steps

1. Override SessionID to a known-missing id.
2. Set OutDir.

```go
import (
	"path/filepath"
	"testing"
)

const unknownDryRunSessionID = "019f283a-eeee-7eee-eeee-eeeeeeeeee99"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = unknownDryRunSessionID
	req.OutDir = filepath.Join(req.TempDir, "dry-run-unknown-out")
	req.ArchivePath = ""
	writeActiveSessions(t, req.GrokHome /* none */)
	return nil
}
```
