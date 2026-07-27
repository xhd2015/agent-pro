# Scenario

**Feature**: `codex-tty` streams Codex rollout JSONL instead of waiting for scrollback fallback

```
fake Codex TUI prints resume UUID in scrollback
  -> agent-run discovers <codex-home>/sessions/YYYY/MM/DD/rollout-*-<uuid>.jsonl
  -> rollout JSONL records become assistant stdout/events while PTY runs
```

## Preconditions

- The fake Codex home is isolated under the test temp dir.
- The fake TUI prints `codex resume <uuid>` so discovery must use terminal scrollback.
- Leaves decide whether the matching rollout transcript exists.

## Steps

1. Set common `agent-run run --agent-runner codex-tty "run ls"` args.
2. Set an isolated `CODEX_HOME` and matching fake Codex session UUID.
3. Use `codex-jsonl-stream-probe` mode so leaves can observe stdout while the PTY is alive.

## Context

- This branch covers the structured Codex transcript source only; existing `run/`
  leaves continue to cover generic PTY lifecycle, registry, attach, and fallback behavior.

```go
import (
	"fmt"
	"testing"
	"github.com/xhd2015/doctest/session"
)

const codexJSONLSessionID = "019f20fd-8569-7910-ab0b-9d898d66e3e6"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Args = []string{"run", "--agent-runner", "codex-tty", "run ls"}
	req.CodexHome = filepath.Join(req.TempDir, ".codex")
	req.CodexTranscriptSessionID = codexJSONLSessionID
	req.Env = withoutEnvKey(req.Env, "FAKE_CODEX_SESSION_ID")
	req.Env = append(req.Env, "FAKE_CODEX_SESSION_ID="+codexJSONLSessionID)
	req.Mode = "codex-jsonl-stream-probe"
	req.ExecTimeout = 15 * time.Second
	req.StreamProbeTimeout = 8 * time.Second
	return nil
}

func fakeTUINoResumeUntilLate(seconds int) string {
	return fmt.Sprintf(`sh -c 'printf "CODEX_TTY_BANNER\nCodex › "; read line; sleep %d; printf "To continue this session, run codex resume %%s\n" "$FAKE_CODEX_SESSION_ID"'`, seconds)
}
```
