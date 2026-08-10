# Scenario

**Feature**: `agent-run takeover` P2 lifecycle — Grok provider session state matrix

```
# Most significant factor: provider existence → managed/liveness → dry-run

missing provider session under GROK_HOME
  -> non-zero; not found

provider present + already managed (registry | process ancestry)
  -> exit 0; warning:; no kill; no iTerm

provider present + takeable
  not running + unmapped -> import-style create + iTerm ForceNew
  not running + mapped   -> resume existing agent-run session + iTerm ForceNew
  running native         -> kill recorded + iTerm ForceNew
  running native + dry-run -> plan only; no kill; no iTerm; no meta create
```

## Preconditions

- Mode handle (takeover root).
- Isolated `AGENT_RUN_HOME` + `GROK_HOME`.
- Injectable hooks via `AGENT_RUN_TAKEOVER_TEST_HOOKS` + `KOOL_ITERM2_*` (see parent SETUP).
- Grok-first: leaves pass `--grok` (or equivalent runner) and a provider UUID.
- No Codex action-body scenarios in P2.

## Steps

1. Leaf seeds Grok / meta / registry / hooks as required.
2. Leaf enables iTerm script capture when open path is expected (or for negative assert).
3. Args = `takeover --grok […] <provider-uuid>`.
4. Assert exit code, messages, kill log, iTerm script, meta side effects.

## Context

Parameter ranking (most → least significant):

1. **Provider session present** under GROK_HOME
2. **Already managed** vs takeable (registry live / process under agent-run)
3. **Liveness** of native grok PIDs (running vs not running)
4. **Mapping** of provider UUID in agent-run meta (import vs resume)
5. **Dry-run** vs acting
