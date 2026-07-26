# Scenario

**Feature**: `LLM_MOCK_RUN_FLAGS` prepends default run flags before argv parsing (GOFLAGS-style)

```
# env tokens prepended, then argv parsed
LLM_MOCK_RUN_FLAGS=--log-events f.jsonl -> llm-mock run grok -> ParseRunFlags -> orchestrator
LLM_MOCK_RUN_FLAGS=--log-events f.jsonl -> llm-mock-run-grok -> ParseRunFlags -> orchestrator

# CLI duplicate overrides env (last-wins)
LLM_MOCK_RUN_FLAGS=--log-events a.jsonl + llm-mock run --log-events b.jsonl grok -> only b.jsonl
```

## Preconditions

- `LLM_MOCK_RUN_FLAGS` is tokenized with simple whitespace split (`strings.Fields`); no quoting.
- Unset or empty `LLM_MOCK_RUN_FLAGS` is a no-op (covered by existing `log-events/` / `mock-events-preset/` leaves).
- Applies to `llm-mock run grok` and `llm-mock-run-grok`; shortcut has no run-flag argv.
- Shared `ParseRunFlags(PrependRunFlagsFromEnv(args))` path for both entry points.
- `Request.OmitCLIRunFlags` — when true, `Run` sets env only and does not pass
  `--log-events` / `--log-http` / `--mock-events-preset` on subcommand argv.

## Steps

1. Grouping `Setup` documents env prepend contract and `OmitCLIRunFlags` pattern.
2. Leaf `Setup` sets `RunFlagsEnv`, optional CLI run flags, entry point (`UseShortcut`), and fake grok profile.
3. `Run` exports `LLM_MOCK_RUN_FLAGS` when `Request.RunFlagsEnv` is non-empty.
4. Leaf `Assert` checks AgentEvent JSONL, CLI override precedence, or preset catalog without grok.

## Context

- `Request.RunFlagsEnv` — full value for `LLM_MOCK_RUN_FLAGS` (e.g. `--log-events /tmp/x.jsonl`).
- `Request.LogEventsPath` — assertion read path; also passed on CLI unless `OmitCLIRunFlags` is true.
- `Request.MockEventsPreset` — CLI preset when not omitted; env may supply `--mock-events-preset=list`.
- `parseAgentEventMaps` — reused from root `SETUP.md` for log-events leaves.

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