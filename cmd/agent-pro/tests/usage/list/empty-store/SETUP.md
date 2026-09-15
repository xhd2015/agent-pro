# Scenario

**Feature**: listing a store with no snapshots yet reads as "nothing yet"

```
usage list      # before the first collect ran
stdout: no usage snapshots under <path>
exit 0
```

## Preconditions

- `AGENT_PRO_HOME` holds no `usages` tree at all.
- A first-time user runs `usage list` after installing; that must be a normal empty
  answer, not an error.

## Steps

1. Keep the store empty: no Setup is needed.
