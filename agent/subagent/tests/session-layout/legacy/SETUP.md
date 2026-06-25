# Scenario

**Feature**: zero-value SessionLayout preserves legacy nested session layout

```
# Dir unset → subagent creates date-nested sess_* under HOME
test -> subagent.Run(SessionLayout{}) -> ~/.agent-pro/subagent/<role>/sessions/<date>/sess_*
```

## Preconditions

- `SessionLayout` is zero value (`Dir` unset).
- `HOME` points at isolated temp directory from root setup.

## Steps

1. Leave `SessionLayout` zeroed and do not pre-create a flat session dir.
2. Assign legacy session id and fake-codex mock config.

## Context

- Role name in harness: `layout-test`
- Legacy session id: `gen_layout_legacy_test`

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/subagent"
)

const legacySessionID = "gen_layout_legacy_test"

func configureLegacyBase(t *testing.T, req *Request) {
	t.Helper()
	req.SessionID = legacySessionID
	req.Layout = subagent.SessionLayout{}
	req.SessionDir = ""

	mockPath := filepath.Join(req.TempDir, "legacy-mock.json")
	writeFile(t, mockPath, minimalMockConfig("inner_legacy_sess"))
	req.MockConfigPath = mockPath
}

func Setup(t *testing.T, req *Request) error {
	configureLegacyBase(t, req)
	if req.HomeDir != "" {
		os.Setenv("HOME", req.HomeDir)
	}
	return nil
}```
