# Scenario

**Feature**: FormatSessionFilesJSON emits a JSON array without ANSI

```
multi-file fixture
-> ListSessionFiles -> FormatSessionFilesJSON
-> JSON array of {name, size, ...}
```

## Preconditions

- Multi-file fixture.
- `req.Format = "json"`.

## Steps

1. Seed three artifacts.
2. Set Format to json.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionID = fixtureFilesSessionID
	writeFilesSession(t, req.GrokHome, req.SessionID, fixtureFilesCWD, defaultMultiFileBodies(req.SessionID))
	req.Format = "json"
	return nil
}
```
