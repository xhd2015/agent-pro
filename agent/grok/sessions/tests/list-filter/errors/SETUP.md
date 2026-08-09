# Scenario

**Feature**: ListWithOptions rejects invalid option combinations

```
# RecentSet with non-positive Recent -> error
# GrepSet with empty Grep -> error
# MainAgent && SubAgent -> error (mutually exclusive)
ListWithOptions(bad opts) -> err; no panic
```

## Preconditions

- Library validates opts before or during filter application.
- Fixtures may be empty; validation should not depend on disk contents.

## Steps

1. Set invalid flag combination on Request.
2. Assert resp.Err is non-nil and message is informative.
