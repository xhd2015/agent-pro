# Scenario

**Feature**: lazy per-UUID on-disk cache under index/by-runner-session

```
cold/stale Find|List
  -> ListSessions scan
  -> write index/by-runner-session/<uuid>.json for each non-empty rsid
  -> set .gen = index/generation
warm (gens equal)
  -> read UUID file or treat missing file as no matches (no scan)
write (Create / Update rsid / ClearAll)
  -> bump generation; next lookup rebuilds
```

## Preconditions

- Cache is a hint; meta.json remains authoritative on rebuild.
- Leaves drive warm + mutate via `req.WarmQueryID` / `req.Mutate`, not ad-hoc switches in Run.
- Asserts use `resp.CacheAfter` and optional `resp.WarmGen`.
