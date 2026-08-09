# Scenario

**Feature**: MainAgent=true keeps only main-agent class

```
MainAgent=true
  -> plain main + fork + empty-kind-no-parent kept
  -> subagent / resume / sub-fork / empty+parent dropped
```

## Preconditions

- MainAgent set; SubAgent and Forked false.
- Fixtures cover every classification cell relevant to main class.

## Steps

1. Write main-class and sub-class fixtures with distinct last_active.
2. MainAgent=true, high Limit.
3. Assert only main-class IDs, newest-first, with Kind tokens.
