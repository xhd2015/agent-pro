# Scenario

**Feature**: `llm-mock run --mock-events-preset=<name> codex` passes preset to background mock server

```
# catalog without codex/mock
llm-mock run --mock-events-preset=list -> stdout catalog, exit 0

# orchestrator pass-through (user example preset)
llm-mock run --mock-events-preset=think-tool-message codex
orchestrator -> mock server --mock-events-preset think-tool-message
fake codex -> curl mock -> preset genQueue dequeue order
```

## Preconditions

- `--mock-events-preset` is a `llm-mock run` flag (`lessflags` with `StopOnFirstArg()`).
- `list` exits 0 without starting mock server or codex (no `CODEX_HOME=` announcement).
- Preset flag passes through to `startMockServer` as `--mock-events-preset`.
- Tokens after `codex` remain codex argv unchanged (same as `--log-events` / `--log-http`).

## Steps

1. Grouping `Setup` documents orchestrator preset contract; leaves set `MockEventsPreset`, `ListOnly`, fake codex profile.
2. `Run` passes `--mock-events-preset` before optional `codex` subcommand.
3. Leaf `Assert` checks catalog stdout or agent-events order from `think-tool-message` preset.

## Context

- `Request.MockEventsPreset` — preset name or `list`.
- `Request.ListOnly` — when true, `Run` invokes `llm-mock run --mock-events-preset=list` without `codex`.
- `Request.LogEventsPath` — optional; pass-through leaf sets this to prove preset dequeue via AgentEvent JSONL.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.UseShortcut = false
	return nil
}
```