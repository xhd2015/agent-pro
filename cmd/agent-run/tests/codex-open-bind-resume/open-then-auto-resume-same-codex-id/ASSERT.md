---
label: e2e,codex
explanation: >-
  codex-tty --open binds meta.runner_session_id; auto-send-or-resume after end
  keeps the same Codex id (Fix A for SeaTalk local-bot resume).
---

## Expected

1. Open exit 0 (or process completed).
2. After open: at least one Codex rollout exists under isolated CODEX_HOME
   (environment smoke — proves real codex + mock ran).
3. After open: `meta.runner_session_id` is **non-empty** and matches a rollout uuid.
4. After auto-send-or-resume: `meta.runner_session_id` is **unchanged**.
5. After auto-send-or-resume: Codex rollout uuid set has **size 1** (same id),
   not a second conversation.

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

	// Environment smoke: open must have produced a real codex session.
	if len(resp.CodexIDsAfterOpen) == 0 {
		t.Fatalf("no codex rollout after open (env/mock broken?)\nopen stderr:\n%s\nstatus:\n%s",
			resp.Open.Stderr, resp.StatusAfterOpen)
	}
	if resp.Open.ExitCode != 0 && resp.Open.Err != nil {
		t.Fatalf("open failed: exit=%d err=%v\nstderr:\n%s", resp.Open.ExitCode, resp.Open.Err, resp.Open.Stderr)
	}

	// Gate A: bind on open (Fix A) — currently FAILS (unbound).
	if strings.TrimSpace(resp.RunnerSessionIDAfterOpen) == "" {
		t.Fatalf("after open: meta.runner_session_id empty (unbound)\n"+
			"codex rollouts: %v\nstatus:\n%s\nmeta: %v\n"+
			"open stderr:\n%s",
			resp.CodexIDsAfterOpen, resp.StatusAfterOpen, resp.MetaAfterOpen, resp.Open.Stderr)
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

	// Gate B: auto-send-or-resume keeps same codex id — currently FAILS (second rollout).
	if strings.TrimSpace(resp.RunnerSessionIDAfterResume) == "" {
		t.Fatalf("after auto-send-or-resume: meta.runner_session_id still empty\n"+
			"codex after open: %v\ncodex after resume: %v\nstatus:\n%s",
			resp.CodexIDsAfterOpen, resp.CodexIDsAfterResume, resp.StatusAfterResume)
	}
	if resp.RunnerSessionIDAfterResume != openID {
		t.Fatalf("runner_session_id changed: open=%s resume=%s\ncodex after open: %v\nafter resume: %v",
			openID, resp.RunnerSessionIDAfterResume, resp.CodexIDsAfterOpen, resp.CodexIDsAfterResume)
	}
	if len(resp.CodexIDsAfterResume) != 1 {
		t.Fatalf("expected exactly 1 codex rollout after resume (same session); got %d ids %v\n"+
			"(second id means ModeRun opened a new Codex conversation)",
			len(resp.CodexIDsAfterResume), resp.CodexIDsAfterResume)
	}
	if resp.CodexIDsAfterResume[0] != openID {
		t.Fatalf("codex rollout id mismatch: want %s got %v", openID, resp.CodexIDsAfterResume)
	}
}
```
