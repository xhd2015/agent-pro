# Scenario

**Feature**: two concurrent codex-tty `--open` on the **same** workspace +
**same** CODEX_HOME bind **distinct** prompt-matched `runner_session_id`s.

```
agent-run run --open --session-id concurrent-A --dir <ws> ... -- "QUESTION_A_combined_checkout"
agent-run run --open --session-id concurrent-B --dir <ws> ... -- "QUESTION_B_item_level_refid"
  (parallel; shared AGENT_RUN_HOME + CODEX_HOME)

  -> meta(A).runner_session_id != meta(B).runner_session_id
  -> each id's first real user prompt matches that open's prompt
```

Crime-scene convert (eval duplicate Agent Answer root cause). Expect **PASS**
after product discovery is prompt-matched / fail-closed under concurrency.

## Steps

1. Root Setup builds binaries + isolates homes.
2. Run opens A and B concurrently.
3. Assert both bound, distinct ids, prompt-correct rollouts.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SessionIDA = "concurrent-A"
	req.SessionIDB = "concurrent-B"
	req.PromptA = "QUESTION_A_combined_checkout"
	req.PromptB = "QUESTION_B_item_level_refid"
	return nil
}
```
