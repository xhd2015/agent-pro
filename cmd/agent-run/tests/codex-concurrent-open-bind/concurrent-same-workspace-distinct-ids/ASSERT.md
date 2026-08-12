---
label: e2e,codex
explanation: >-
  concurrent codex-tty --open on same workspace must bind distinct
  runner_session_ids matched to each open prompt (crime-scene fix gate).
---

## Expected

1. Both opens complete (exit 0 preferred; environment may still produce rollouts).
2. At least **two** Codex rollouts under isolated CODEX_HOME (two real threads).
3. `meta(concurrent-A).runner_session_id` and `meta(concurrent-B).runner_session_id`
   are both **non-empty**.
4. The two runner_session_ids are **different** (primary concurrency gate).
5. First real user prompt in each bound rollout matches that session's open prompt
   (A → QUESTION_A…, B → QUESTION_B…; skip `<environment_context>`).

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

	// Environment smoke: concurrent opens should create two codex threads.
	if len(resp.CodexIDs) < 2 {
		t.Fatalf("expected >=2 codex rollouts after concurrent open; got %d ids %v\n"+
			"openA stderr:\n%s\nopenB stderr:\n%s",
			len(resp.CodexIDs), resp.CodexIDs, resp.OpenA.Stderr, resp.OpenB.Stderr)
	}

	idA := strings.TrimSpace(resp.RunnerSessionIDA)
	idB := strings.TrimSpace(resp.RunnerSessionIDB)
	if idA == "" {
		t.Fatalf("session A unbound: meta.runner_session_id empty\n"+
			"codex ids: %v\nmetaA: %v\nopenA stderr:\n%s",
			resp.CodexIDs, resp.MetaA, resp.OpenA.Stderr)
	}
	if idB == "" {
		t.Fatalf("session B unbound: meta.runner_session_id empty\n"+
			"codex ids: %v\nmetaB: %v\nopenB stderr:\n%s",
			resp.CodexIDs, resp.MetaB, resp.OpenB.Stderr)
	}

	// Primary gate (crime scene): must not share one codex thread.
	if idA == idB {
		t.Fatalf("concurrent opens bound the SAME runner_session_id=%s\n"+
			"(desired: distinct ids matched to each prompt)\n"+
			"promptA=%q promptB=%q\ncodex rollouts: %v\npromptById: %v\n"+
			"openA stderr:\n%s\nopenB stderr:\n%s",
			idA, req.PromptA, req.PromptB, resp.CodexIDs, resp.PromptByCodexID,
			resp.OpenA.Stderr, resp.OpenB.Stderr)
	}

	// Correctness: each bind points at the rollout that received that open prompt.
	gotA := strings.TrimSpace(resp.PromptByCodexID[idA])
	gotB := strings.TrimSpace(resp.PromptByCodexID[idB])
	if gotA != strings.TrimSpace(req.PromptA) {
		t.Fatalf("session A bound id=%s has first real user prompt %q; want %q\n"+
			"promptById: %v", idA, gotA, req.PromptA, resp.PromptByCodexID)
	}
	if gotB != strings.TrimSpace(req.PromptB) {
		t.Fatalf("session B bound id=%s has first real user prompt %q; want %q\n"+
			"promptById: %v", idB, gotB, req.PromptB, resp.PromptByCodexID)
	}
}
```
