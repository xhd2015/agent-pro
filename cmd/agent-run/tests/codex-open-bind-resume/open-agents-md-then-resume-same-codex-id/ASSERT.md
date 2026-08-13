---
label: e2e,codex
explanation: >-
  Workspace AGENTS.md must not prevent bind: after --open,
  runner_session_id is set; auto-send-or-resume keeps the same Codex uuid
  (no second conversation).
---

## Expected

1. Open produced at least one Codex rollout (env smoke).
2. After open: `meta.runner_session_id` non-empty and in that rollout set.
3. After auto-send-or-resume: same `runner_session_id`.
4. After auto-send-or-resume: exactly one Codex rollout uuid.

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatalf("run error: %v", err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if len(resp.CodexIDsAfterOpen) == 0 {
		t.Fatalf("no codex rollout after open (env/mock broken?)\nopen stderr:\n%s\nstatus:\n%s",
			resp.Open.Stderr, resp.StatusAfterOpen)
	}
	if resp.Open.ExitCode != 0 && resp.Open.Err != nil {
		t.Fatalf("open failed: exit=%d err=%v\nstderr:\n%s", resp.Open.ExitCode, resp.Open.Err, resp.Open.Stderr)
	}

	if strings.TrimSpace(resp.RunnerSessionIDAfterOpen) == "" {
		t.Fatalf("after open: meta.runner_session_id empty (AGENTS.md first user message hid inject)\n"+
			"codex rollouts: %v\nstatus:\n%s\nopen stderr:\n%s",
			resp.CodexIDsAfterOpen, resp.StatusAfterOpen, resp.Open.Stderr)
	}
	openID := resp.RunnerSessionIDAfterOpen
	found := false
	for _, id := range resp.CodexIDsAfterOpen {
		if id == openID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("after open: runner_session_id=%s not in rollouts %v", openID, resp.CodexIDsAfterOpen)
	}

	if strings.TrimSpace(resp.RunnerSessionIDAfterResume) == "" {
		t.Fatalf("after auto-send-or-resume: still unbound\n"+
			"codex after open: %v\nafter resume: %v\nstatus:\n%s",
			resp.CodexIDsAfterOpen, resp.CodexIDsAfterResume, resp.StatusAfterResume)
	}
	if resp.RunnerSessionIDAfterResume != openID {
		t.Fatalf("runner_session_id changed: open=%s resume=%s\nrollouts open=%v resume=%v",
			openID, resp.RunnerSessionIDAfterResume, resp.CodexIDsAfterOpen, resp.CodexIDsAfterResume)
	}
	if len(resp.CodexIDsAfterResume) != 1 {
		t.Fatalf("expected 1 codex rollout after resume; got %d %v (second id = ModeRun new conversation)",
			len(resp.CodexIDsAfterResume), resp.CodexIDsAfterResume)
	}
	if resp.CodexIDsAfterResume[0] != openID {
		t.Fatalf("codex rollout id mismatch: want %s got %v", openID, resp.CodexIDsAfterResume)
	}
}
```
