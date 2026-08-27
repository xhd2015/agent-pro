# Scenario

**Feature**: Grok composer placeholder `Build anything` is empty, not occupied

```
 │ ❯ Build anything                                                                            │
  -> DetectInputBox
  -> empty
```

Live idle chrome shows placeholder text inside the boxed `❯` composer; it is not a user draft.

## Steps

1. Load post-turn Recap fixture that includes the `Build anything` placeholder.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	root := filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", "..", "..", ".."))
	path := filepath.Join(root, "pkgs", "agenttty", "testdata", "grok-writable",
		"grok-after_recap-expand-thinking-idle-01a03d6f.txt")
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	req.Scrollback = string(b)
	req.Fixture = ""
	return nil
}
```
