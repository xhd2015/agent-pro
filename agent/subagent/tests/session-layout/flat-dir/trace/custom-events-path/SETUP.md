# Scenario

**Feature**: traceSession reads custom EventsPath after Run

```
# EventsPath outside Dir; Run then traceSession
trace stdout shows Events: N lines from external path
```

## Preconditions

- Custom `EventsPath` configured on `SessionLayout`.

## Steps

1. Create flat session dir and external events path (same as custom-paths leaf).
2. Run subagent via doctest `Run`.
3. Assert calls `runTrace` with same layout.

## Context

- Inner session id: `inner_trace_sess`

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/subagent"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionDir = filepath.Join(req.TempDir, "trace-session")
	if err := os.MkdirAll(req.SessionDir, 0755); err != nil {
		return err
	}

	externalDir := filepath.Join(req.TempDir, "trace-external-events")
	if err := os.MkdirAll(externalDir, 0755); err != nil {
		return err
	}
	req.CustomEventsPath = filepath.Join(externalDir, "events.jsonl")
	req.SessionID = flatSessionID
	req.Layout = subagent.SessionLayout{
		Dir:              req.SessionDir,
		EventsPath:       req.CustomEventsPath,
		MessagesPath:     "",
		QuestionsEnabled: false,
		ProgressEnabled:  false,
	}

	mockPath := filepath.Join(req.TempDir, "trace-mock.json")
	writeFile(t, mockPath, minimalMockConfig("inner_trace_sess"))
	req.MockConfigPath = mockPath
	return nil
}```
