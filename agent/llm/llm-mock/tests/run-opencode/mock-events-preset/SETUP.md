# Scenario

**Feature**: `llm-mock run --mock-events-preset=<name> opencode` passes preset to background mock server

```
# catalog without opencode/mock
llm-mock run --mock-events-preset=list -> stdout catalog, exit 0

# orchestrator pass-through (simple preset)
llm-mock run --mock-events-preset=simple opencode
orchestrator -> mock server --mock-events-preset simple
fake opencode -> curl mock -> preset genQueue dequeue order
```

## Preconditions

- `--mock-events-preset` is a `llm-mock run` flag (`lessflags` with `StopOnFirstArg()`).
- `list` exits 0 without starting mock server or opencode (no `OPENCODE_CONFIG_DIR=` announcement).
- Preset flag passes through to `startMockServer` as `--mock-events-preset`.
- Tokens after `opencode` remain opencode argv unchanged (same as `--log-events` / `--log-http`).

## Steps

1. Grouping `Setup` documents orchestrator preset contract; leaves set `MockEventsPreset`, `ListOnly`, fake opencode profile.
2. `Run` passes `--mock-events-preset` before optional `opencode` subcommand.
3. Leaf `Assert` checks catalog stdout or agent-events from `simple` preset.

## Context

- `Request.MockEventsPreset` — preset name or `list`.
- `Request.ListOnly` — when true, `Run` invokes `llm-mock run --mock-events-preset=list` without `opencode`.
- `Request.LogEventsPath` — optional; pass-through leaf sets this to prove preset dequeue via AgentEvent JSONL.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.UseShortcut = false
	return nil
}
```