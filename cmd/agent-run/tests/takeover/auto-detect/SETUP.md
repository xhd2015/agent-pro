# Scenario

**Feature**: `agent-run takeover <session-id>` auto-detects provider when runner flags are omitted

```
# no --grok / --codex / --agent-runner
takeover <uuid>
  -> Find GROK_HOME; Find CODEX_HOME
  -> exactly one -> executeTakeover with grok-tty | codex-tty
  -> both -> ambiguous (mention both)
  -> neither -> not found / cannot resolve provider

# explicit flags still skip auto-detect (covered by P2/P3; not retested here)
```

## Preconditions

- Mode handle; isolated `AGENT_RUN_HOME` + `GROK_HOME` + `CODEX_HOME` from root.
- Leaves do **not** pass `--grok`, `--codex`, or `--agent-runner`.
- Success leaves use not-running import fixtures (empty hooks + iTerm capture).
- Parallel-safe: `req.Env` only.

## Steps

1. Seed zero/one/both provider homes for the UUID under test.
2. Empty hooks (+ iTerm when success path expected).
3. Args = `takeover <uuid>` only.
4. Assert resolve outcome (success lifecycle or error wording).

## Context

Parameter ranking:

1. **Hits under homes** — grok-only | codex-only | both | neither
2. **Outcome** — run lifecycle vs ambiguous vs not found

Today empty runner errors with `takeover requires --grok, --codex, or --agent-runner`
before Find → all leaves **RED** until auto-detect lands.
