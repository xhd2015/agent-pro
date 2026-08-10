# Scenario

**Feature**: `takeover` early argument and flag validation

```
agent-run takeover -> missing session-id; non-zero
agent-run takeover "   " -> empty session-id after trim; non-zero
agent-run takeover --grok --codex <id> -> mutually exclusive
agent-run takeover --grok --agent-runner codex-tty <id> -> mismatch
agent-run takeover --codex --agent-runner grok-tty <id> -> mismatch
agent-run takeover --not-a-real-flag <id> -> unknown flag; non-zero
```

## Preconditions

- Empty `AGENT_RUN_HOME` (no registry / no sessions).
- Mode handle (set by takeover root Setup).
- No provider process scan / kill / iTerm (P1 only).

## Steps

1. Leaf sets Args for the validation case.
2. Run Handle.
3. Assert non-zero exit and error text (not merely `unknown command: takeover`).
