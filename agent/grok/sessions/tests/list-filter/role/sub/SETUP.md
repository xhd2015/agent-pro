# Scenario

**Feature**: SubAgent=true keeps only sub-agent class

```
SubAgent=true
  -> subagent / subagent_resume / subagent_fork / empty-kind+parent kept
  -> plain main + fork + empty-kind-no-parent dropped
```

## Preconditions

- SubAgent set; MainAgent and Forked false.

## Steps

1. Seed same multi-kind fixture set as main-class leaf.
2. SubAgent=true.
3. Assert only sub-class IDs and Kind tokens.
