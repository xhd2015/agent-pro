# Scenario

**Feature**: fs/list returns dirs and files with type field (A6)

```
fixture: <root>/subdir/ + <root>/note.txt
GET /fs/list?path=<root>
  -> entries include subdir type=dir, note.txt type=file
  -> parent path present when applicable
```

## Preconditions

- Fixture from `makeFixtureTree`.
- Expect RED until endpoint exists.

## Steps

1. Build fixture; start web; GET list.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Scenario = "dirs-and-files-typed"
	root := makeFixtureTree(t, req)
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}
	req.HTTPSteps = []HTTPStep{
		{Name: "list", Method: "GET", Path: fsListPath(root)},
	}
	return nil
}
```
