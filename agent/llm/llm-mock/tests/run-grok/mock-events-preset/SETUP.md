# Scenario

**Feature**: `llm-mock run --mock-events-preset=<name> grok` passes preset to background mock server

```
# catalog without grok/mock
llm-mock run --mock-events-preset=list -> stdout catalog, exit 0

# orchestrator pass-through
llm-mock run --mock-events-preset=think-message grok
orchestrator -> mock server --mock-events-preset think-message
fake grok -> curl mock -> preset genQueue dequeue order
```

## Preconditions

- `--mock-events-preset` is a `llm-mock run` flag (`lessflags` with `StopOnFirstArg()`).
- `list` exits 0 without starting mock server or grok (no `GROK_HOME=` announcement).
- Unknown preset errors before mock/grok start (not covered here — see server tree `unknown-preset`).
- Preset flag passes through to `startMockServer` as `--mock-events-preset`.
- Tokens after `grok` remain grok argv unchanged (same as `--log-events` / `--log-http`).

## Steps

1. Grouping `Setup` documents orchestrator preset contract; leaves set `MockEventsPreset`, `ListOnly`, fake grok profile.
2. `Run` passes `--mock-events-preset` before optional `grok` subcommand.
3. Leaf `Assert` checks catalog stdout, agent-events order, or grok argv passthrough.

## Context

- `Request.MockEventsPreset` — preset name or `list`.
- `Request.ListOnly` — when true, `Run` invokes `llm-mock run --mock-events-preset=list` without `grok`.
- `Request.LogEventsPath` — optional; pass-through leaves may set this to prove preset dequeue via AgentEvent JSONL.

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