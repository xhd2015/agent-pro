# Scenario

**Feature**: missing Codex rollout transcript keeps cleaned scrollback fallback

```
fake Codex TUI prints resume UUID
  -> no matching rollout JSONL exists under CODEX_HOME
  -> agent-run falls back to cleaned terminal scrollback
```

## Preconditions

- The Codex home exists but contains no matching rollout transcript.

## Steps

1. Configure the fake TUI to print a resume line and a fallback assistant line.
2. Do not create any rollout JSONL file.
3. Assert the fallback line reaches stdout.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

const codexScrollbackFallbackText = "SCROLLBACK_FALLBACK_WITHOUT_TRANSCRIPT"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.CodexTTYCommand = fakeTUICodexResumeWithFallback(codexScrollbackFallbackText)
	req.Env = withoutEnvKey(req.Env, "FAKE_CODEX_FALLBACK_TEXT")
	req.Env = append(req.Env, "FAKE_CODEX_FALLBACK_TEXT="+codexScrollbackFallbackText)
	req.StreamProbeSubstring = codexScrollbackFallbackText
	return nil
}
```
