# Scenario

**Feature**: flat SessionLayout.Dir uses caller-provided directory without date nesting

```
# caller MkdirAll flat dir; subagent skips sess_* auto-create
test -> subagent.Run(SessionLayout{Dir: flat}) -> artifacts directly under flat dir
```

## Preconditions

- `SessionLayout.Dir` points to an existing empty directory.

## Steps

1. Create flat session directory under temp dir.
2. Set default layout with questions and progress enabled.
3. Write fake-codex mock config.

## Context

- Descendant leaves narrow `MessagesPath`, feature flags, or pre-created meta.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/subagent"
)

const flatSessionID = "gen_layout_flat_test"

func configureFlatDirBase(t *testing.T, req *Request) {
	t.Helper()
	req.SessionDir = filepath.Join(req.TempDir, "flat-session")
	if err := os.MkdirAll(req.SessionDir, 0755); err != nil {
		t.Fatal(err)
	}
	req.SessionID = flatSessionID
	req.Layout = subagent.SessionLayout{
		Dir:              req.SessionDir,
		MessagesPath:     "messages.jsonl",
		QuestionsEnabled: true,
		ProgressEnabled:  true,
	}
	mockPath := filepath.Join(req.TempDir, "layout-mock.json")
	writeFile(t, mockPath, minimalMockConfig("inner_layout_sess"))
	req.MockConfigPath = mockPath
}

func Setup(t *testing.T, req *Request) error {
	configureFlatDirBase(t, req)
	return nil
}
```
