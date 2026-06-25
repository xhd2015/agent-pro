# Scenario

**Feature**: flat Dir with zero feature flags skips questions and progress

```
# task-hub wiring: Dir set, QuestionsEnabled/ProgressEnabled false
subagent.Run(flat, flags off) -> no questions/ or progress/
```

## Preconditions

- `SessionLayout.Dir` set; `QuestionsEnabled` and `ProgressEnabled` left at zero.

## Steps

1. Create flat session dir with only `Dir` set (no explicit flag enablement).

## Context

- Mirrors task-hub integration where optional features are opt-in.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/subagent"
)

func configureFlatDirFlagsOff(t *testing.T, req *Request) {
	t.Helper()
	req.SessionDir = filepath.Join(req.TempDir, "flags-off-session")
	if err := os.MkdirAll(req.SessionDir, 0755); err != nil {
		t.Fatal(err)
	}
	req.SessionID = flatSessionID
	req.Layout = subagent.SessionLayout{
		Dir: req.SessionDir,
	}
	mockPath := filepath.Join(req.TempDir, "flags-off-mock.json")
	writeFile(t, mockPath, minimalMockConfig("inner_flags_off_sess"))
	req.MockConfigPath = mockPath
}

func Setup(t *testing.T, req *Request) error {
	configureFlatDirFlagsOff(t, req)
	return nil
}```
