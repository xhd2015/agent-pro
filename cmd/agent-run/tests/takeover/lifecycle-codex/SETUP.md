# Scenario

**Feature**: `agent-run takeover --codex` P3 lifecycle — Codex provider session state matrix

```
# MECE sibling of lifecycle/ (Grok): same factors, provider kind = codex

missing provider session under CODEX_HOME
  -> non-zero; not found

provider present + already managed (registry | process ancestry)
  -> exit 0; warning:; no kill; no iTerm

provider present + takeable
  not running + unmapped -> import meta runner=codex-tty + iTerm ForceNew
  not running + mapped   -> resume existing agent-run session + iTerm ForceNew
  running native         -> kill recorded + iTerm ForceNew
  running native + dry-run -> plan only; no kill; no iTerm; no meta create
```

## Preconditions

- Mode handle (takeover root).
- Isolated `AGENT_RUN_HOME` + `CODEX_HOME` (`TempDir/.codex`).
- Injectable hooks via `AGENT_RUN_TAKEOVER_TEST_HOOKS` + `KOOL_ITERM2_*`.
- Codex-first: leaves pass `--codex` and a provider UUID.
- Rollout seed: `CODEX_HOME/sessions/YYYY/MM/DD/rollout-*-<uuid>.jsonl` with `session_meta`.
- Process cmd basename `codex`; open path under `…/.codex/sessions/…/rollout-…-<uuid>.jsonl`.
- Registry: `codex-tty-registry/` for already-managed-registry.
- Mapped meta: `runner=codex-tty`, `runner_session_id` = provider UUID.

## Steps

1. Leaf seeds Codex rollout / meta / registry / hooks as required.
2. Leaf enables iTerm script capture when open path is expected (or for negative assert).
3. Args = `takeover --codex […] <provider-uuid>`.
4. Assert exit code, messages, kill log, iTerm script, meta side effects.

## Context

Parameter ranking (same as Grok P2):

1. **Provider session present** under CODEX_HOME
2. **Already managed** vs takeable
3. **Liveness** of native codex PIDs
4. **Mapping** of provider UUID in agent-run meta
5. **Dry-run** vs acting

Product today returns `takeover: codex support is not implemented yet` → all leaves **RED**.
