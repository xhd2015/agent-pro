# Scenario

**Feature**: help surfaces list and document `takeover`

```
agent-run --help -> lists takeover as a command
agent-run takeover --help -> usage, <session-id>, --grok, --codex, --agent-runner, --dry-run
```

## Preconditions

- L2 Mode handle (set by takeover root Setup).
- No live fixtures required.

## Steps

1. Leaf sets `req.Args` for the help invocation.
2. Run Handle in-process.
3. Assert exit 0 and required tokens / trailing newline.
