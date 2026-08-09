# Scenario

**Feature**: ListWithOptions with zero filters behaves like baseline List

```
# no PlaceCWDs / RecentSet / Active / GrepSet
ListWithOptions(opts zero filters) -> all discovered sessions, sort + limit only
```

## Preconditions

- Descendants leave place/recent/active/grep unset (zero value).
- Limit may be set per leaf.

## Steps

1. Ensure sessions root exists (root Setup).
2. Leaf seeds fixtures and Limit / WantFormat as needed.
