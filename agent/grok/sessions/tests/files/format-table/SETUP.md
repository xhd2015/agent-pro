# Scenario

**Feature**: FormatSessionFilesTable prints NAME SIZE MTIME table

```
multi-file fixture
-> ListSessionFiles -> FormatSessionFilesTable
-> header NAME SIZE MTIME; rows include basenames
```

## Preconditions

- Same multi-file fixture as `multi-file/`.
- `req.Format = "table"`.

## Steps

1. Seed three artifacts.
2. Set Format to table.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureFilesSessionID
	writeFilesSession(t, req.GrokHome, req.SessionID, fixtureFilesCWD, defaultMultiFileBodies(req.SessionID))
	req.Format = "table"
	return nil
}
```
