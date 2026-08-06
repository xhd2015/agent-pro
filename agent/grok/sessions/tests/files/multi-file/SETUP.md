# Scenario

**Feature**: ListSessionFiles returns all regular files in the session dir

```
seed summary.json + updates.jsonl + signals.json
-> ListSessionFiles
-> dir = session path; files include all three basenames with Size>0, Path set
```

## Preconditions

- Three non-empty artifacts under the session directory.
- Session id is discoverable via Find (valid `summary.json`).

## Steps

1. Write multi-file fixture with default bodies.
2. Set `SessionID`; Format empty (structured only).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureFilesSessionID
	writeFilesSession(t, req.GrokHome, req.SessionID, fixtureFilesCWD, defaultMultiFileBodies(req.SessionID))
	return nil
}
```
