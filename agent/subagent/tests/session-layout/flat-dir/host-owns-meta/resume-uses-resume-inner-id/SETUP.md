# Scenario

**Feature**: resume uses ResumeInnerSessionID instead of reading meta

```
# host passes opencode id via Options; second Run appends events; meta frozen
first Run -> ResumeInnerSessionID=inner_resume_host -> second Run appends events
```

## Preconditions

- Host `meta.json` without `opencode_session_id`.
- First run establishes `events.jsonl`.
- Second run sets `ResumeInnerSessionID` to first mock inner id.

## Steps

1. Create flat dir and host meta; snapshot bytes.
2. First `Run` via doctest harness.
3. Store second mock; Assert invokes second `invokeRun` with `ResumeInnerSessionID`.

## Context

- Session id: `gen_layout_host_resume_test`
- First inner id: `inner_host_resume_run1` (resume target for second run)

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/subagent"
)

func Setup(t *testing.T, req *Request) error {
	req.SessionDir = filepath.Join(req.TempDir, "host-resume-session")
	if err := os.MkdirAll(req.SessionDir, 0755); err != nil {
		return err
	}
	req.SessionID = "gen_layout_host_resume_test"
	hostMeta := `{
  "id": "20260625-fixture-host-resume",
  "task_title": "host resume",
  "agent_session_id": "gen_layout_host_resume_test",
  "project_name": "fixture-a"
}`
	metaPath := filepath.Join(req.SessionDir, "meta.json")
	writeFile(t, metaPath, hostMeta)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatal(err)
	}
	req.MetaBytesBeforeRun = data

	req.Layout = subagent.SessionLayout{
		Dir:              req.SessionDir,
		MessagesPath:     "",
		QuestionsEnabled: false,
		ProgressEnabled:  false,
	}

	firstMock := filepath.Join(req.TempDir, "host-resume-mock-1.json")
	writeFile(t, firstMock, minimalMockConfig("inner_host_resume_run1"))
	req.MockConfigPath = firstMock

	secondMock := filepath.Join(req.TempDir, "host-resume-mock-2.json")
	writeFile(t, secondMock, `{
  "version": "agent-pro.fake-runner.v1",
  "runner": "fake-codex",
  "session_id": "inner_host_resume_run1",
  "llm_events": [
    {"type": "message", "text": "host resume second run\n"}
  ]
}`)
	req.SecondMockConfigPath = secondMock
	return nil
}```
