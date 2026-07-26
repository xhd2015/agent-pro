# Scenario

**Feature**: second Run with same SessionID appends events and preserves host meta

```
# pre-created meta + first Run + second Run
events.jsonl grows; host fields unchanged; opencode_session_id set
```

## Preconditions

- Host `meta.json` with `agent_session_id` and task-hub fields, no `opencode_session_id`.

## Steps

1. Create flat session dir and write foreign meta before any run.
2. Configure first mock (`inner_resume_run1`) for initial Run (via doctest Run).
3. Store second mock path (`inner_resume_run2`) for Assert to invoke again.

## Context

- Session id: `gen_layout_resume_test`

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/agent-pro/agent/subagent"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SessionDir = filepath.Join(req.TempDir, "resume-session")
	if err := os.MkdirAll(req.SessionDir, 0755); err != nil {
		return err
	}
	req.SessionID = "gen_layout_resume_test"
	req.PreCreateMeta = `{
  "id": "20260625-fixture-resume",
  "task_title": "resume title",
  "agent_session_id": "gen_layout_resume_test",
  "project_name": "fixture-resume"
}`
	writeFile(t, filepath.Join(req.SessionDir, "meta.json"), req.PreCreateMeta)

	req.Layout = subagent.SessionLayout{
		Dir:              req.SessionDir,
		MessagesPath:     "",
		QuestionsEnabled: false,
		ProgressEnabled:  false,
	}

	firstMock := filepath.Join(req.TempDir, "resume-mock-1.json")
	writeFile(t, firstMock, minimalMockConfig("inner_resume_run1"))
	req.MockConfigPath = firstMock

	secondMock := filepath.Join(req.TempDir, "resume-mock-2.json")
	writeFile(t, secondMock, `{
  "version": "agent-pro.fake-runner.v1",
  "runner": "fake-codex",
  "session_id": "inner_resume_run1",
  "llm_events": [
    {"type": "message", "text": "resume second run\n"}
  ]
}`)
	req.SecondMockConfigPath = secondMock
	return nil
}```
