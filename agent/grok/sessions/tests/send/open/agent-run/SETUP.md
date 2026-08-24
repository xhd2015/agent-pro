# Scenario

**Feature**: `--open` prefers agent-run when the Grok id is managed there

```
no iTerm host + --open
  → FindByGrokSessionID / AgentRunOpen hook
  → live: deliver via agent-run (no grok --resume)
  → exited: agent-run resume window
  → --no-agent-run: bare grok --resume
```
