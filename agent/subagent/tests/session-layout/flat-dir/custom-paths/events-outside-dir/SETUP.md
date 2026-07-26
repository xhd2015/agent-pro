# Scenario

**Feature**: custom EventsPath writes events outside SessionLayout.Dir

```
# EventsPath points to sibling dir under temp
subagent.Run -> events at external path, not Dir/events.jsonl
```

## Preconditions

- Flat session dir exists; external events path is pre-created.

## Steps

1. Create flat session dir and external events directory.
2. Set `SessionLayout.EventsPath` to external absolute path.
3. Disable questions/progress for simpler assertions.

## Context

- Inner session id: `inner_custom_events_sess`
- External events: `<temp>/external-events/events.jsonl`

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/subagent"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionDir = filepath.Join(req.TempDir, "custom-paths-session")
	if err := os.MkdirAll(req.SessionDir, 0755); err != nil {
		return err
	}

	externalDir := filepath.Join(req.TempDir, "external-events")
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

	mockPath := filepath.Join(req.TempDir, "custom-events-mock.json")
	writeFile(t, mockPath, minimalMockConfig("inner_custom_events_sess"))
	req.MockConfigPath = mockPath
	return nil
}```
