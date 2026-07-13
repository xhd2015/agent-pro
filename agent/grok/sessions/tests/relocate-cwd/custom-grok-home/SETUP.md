# Scenario

**Feature**: `opts.GrokHome` selects the fixture home exclusively

```
customHome has session under encode(ws-old)
decoyHome has unrelated marker file only
opts.GrokHome = customHome
-> RelocateCWD moves under customHome only; decoyHome untouched
```

## Preconditions

- Root setup already created `req.GrokHome` (unused as the selected home).
- Leaf creates a **custom** home with the real fixture and a **decoy** home.
- `req.OptsGrokHome` points at the custom home (not decoy, not root GrokHome).

## Steps

1. Create `custom-home` and `decoy-home` under temp.
2. Seed session only under custom home; decoy gets a marker file.
3. Create target workspace; set `OptsGrokHome` to custom home.
4. Record paths for asserts.

```go
import (
	"path/filepath"
	"testing"
)

const customHomeSessionID = "019f283a-ffff-7fff-ffff-ffffffffff06"

func Setup(t *testing.T, req *Request) error {
	customHome := filepath.Join(req.TempDir, "custom-grok-home")
	decoyHome := filepath.Join(req.TempDir, "decoy-grok-home")
	mustMkdir(t, filepath.Join(customHome, "sessions"))
	mustMkdir(t, filepath.Join(decoyHome, "sessions"))

	// Decoy marker must remain after RelocateCWD.
	mustWriteFile(t, filepath.Join(decoyHome, "sessions", "decoy-marker.txt"), "decoy-stay\n")
	// Also plant a decoy sqlite so accidental default-home writes would be visible.
	mustWriteFile(t, filepath.Join(decoyHome, "sessions", "session_search.sqlite"), "DECOY-SQLITE\n")

	oldWS := filepath.Join(req.TempDir, "ws-old")
	newWS := filepath.Join(req.TempDir, "ws-new")
	mustMkdir(t, oldWS)
	mustMkdir(t, newWS)

	req.OldCWD = absPath(t, oldWS)
	req.TargetDir = absPath(t, newWS)
	req.SessionID = customHomeSessionID
	req.OptsGrokHome = customHome
	req.DecoyGrokHome = decoyHome
	req.GrokHome = customHome // helpers / asserts use GrokHome as the home under test
	req.UpdatesMarker = `{"type":"init","marker":"custom-home"}` + "\n"
	req.SQLiteMarker = "CUSTOM-HOME-SQLITE-v1"

	req.SessionDir = writeRelocateSession(t, customHome, req.SessionID, req.OldCWD, relocateSessionOpts{
		Title:              "custom home",
		WritePromptContext: true,
		UpdatesBody:        req.UpdatesMarker,
	})
	req.SQLitePath = writeSQLiteMarker(t, customHome, req.SQLiteMarker)
	writeActiveSessions(t, customHome /* none */)
	return nil
}
```
