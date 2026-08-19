# Scenario

**Feature**: finished live Codex prompt (no stub banner) is an idle screen

```
post-turn › / » chrome, no CODEX_TTY_BANNER
  -> DetectScreenStatus
  -> idle
```

## Steps

1. Leaves inject a finished-prompt fixture. Grouping node has no extra Setup.
