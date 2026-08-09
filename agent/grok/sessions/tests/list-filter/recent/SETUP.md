# Scenario

**Feature**: RecentSet keeps sessions with last_active within Now−Recent

```
# Recent window from injectable Now
ListWithOptions(RecentSet, Recent, Now) -> last_active >= Now-Recent
```

## Preconditions

- Descendants set RecentSet=true and a positive Recent duration.
- Now is fixed by root Setup (fixedNow).
- Inclusive lower bound: last_active equal to cutoff is kept.

## Steps

1. Leaf seeds sessions with known last_active offsets from fixedNow.
2. Leaf sets Recent / RecentSet / Limit.
