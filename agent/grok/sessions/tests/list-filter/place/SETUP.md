# Scenario

**Feature**: PlaceCWDs OR-filter sessions by Abs+Clean info.cwd

```
# PlaceCWDs non-empty -> keep sessions whose cwd matches any place entry
ListWithOptions(PlaceCWDs) -> place survivors only
```

## Preconditions

- Descendants set `req.PlaceCWDs` to one or more absolute (or Abs-resolvable) paths.
- No Recent/Active/Grep unless a leaf explicitly needs them (none here).

## Steps

1. Leaf writes sessions under known fixture cwds (cwdA/B/C).
2. Leaf sets PlaceCWDs via `absPath` helpers.
3. Assert ids / emptiness.
