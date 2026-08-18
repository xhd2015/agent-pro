# Scenario

**Feature**: exact-trim grok runner predicate used by Find filter

```
IsGrokRunner(runner)
  -> true only for trim("grok") or trim("grok-tty")
  # not prefix, not case-fold
```

## Preconditions

- Grouping for `IsGrokRunner` table leaves.
- No store / cache required for this branch.

## Steps

1. Leaf sets `req.Op = "is_grok"` and `req.RunnerCases`.
2. `Run` records `IsGrokRunner` for each case.
3. Assert compares booleans to `Want`.
