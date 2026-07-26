# Scenario

**Feature**: fs/list returns dirs before files and includes dot dirs/files (A1–A3)

```
fixture: <root>/.git/ + src/ + .env + a.txt + note.txt
GET /fs/list?path=<root>
  -> entries include .git type=dir and src type=dir
  -> all dirs appear before any file
  -> .env present as type=file
  -> within dirs and within files: case-insensitive name order
```

## Preconditions

- Rich fixture from `makeChooserOptimizeFixture` (dot dirs + regular dirs + dot file + files).
- Expect RED until handleFSList stops skipping `.` names and sorts dirs-first.

## Steps

1. Build fixture; start web; GET list.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Scenario = "dirs-before-files-include-dots"
	root := makeChooserOptimizeFixture(t, req)
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
