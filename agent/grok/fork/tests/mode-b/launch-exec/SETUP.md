# Scenario

**Feature**: Mode B actually execs built llm-mock-run-grok with an isolated hook

```
buildOnce llm-mock-run-grok
opts.Env: GROK_HOME + LLM_MOCK_RUN_GROK_COMMAND (writes marker, exit 0)
fork.Main(["--session-id", id]) -> child exec -> marker exists
```

## Preconditions

- Child env only (no `t.Setenv`). Hook must keep the mock from opening a TUI.

## Steps

1. Build `./agent/llm/llm-mock/llm-mock-run-grok` into the session cache.
2. Set `ExecMock`, `GrokBin`, hook env, `GROK_HOME`.

## Context

- Marker: `$GROK_HOME/fork-hook-ok.txt`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GrokBin = buildMockGrokOnce(t, d)
	req.ExecMock = true
	req.HookMarker = filepath.Join(req.GrokHome, "fork-hook-ok.txt")
	hook := `mkdir -p "$GROK_HOME/hook-sibling" && echo ok > "$GROK_HOME/fork-hook-ok.txt"`
	req.Env = []string{
		"GROK_HOME=" + req.GrokHome,
		"LLM_MOCK_RUN_GROK_COMMAND=" + hook,
	}
	req.Args = []string{"--session-id", fixtureSessionID}
	return nil
}
```
