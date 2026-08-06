# Scenario

**Feature**: list/show EnrichMode — light default, stale catalog, heavy Find fallback

```
# light (default): session_dir + summary.json only; never Find
# stale/off: stored catalog only; Orphaned=false; no FS / no orphan warnings
# heavy: light first, then Find(grokHome, id) if needed
preseed store + optional live summary mutate / wrong session_dir
  -> ListBookmarks | GetBookmark(enrich)
  -> refreshed or snapshot views + Orphaned / warnings
```

## Preconditions

- Homes injected via root Setup (parallel-safe; no Setenv/Chdir).
- Pin path unchanged (still Find once); these leaves only exercise list/show.
- Product EnrichMode not implemented yet — leaves expect RED until implementer.

## Steps

1. Leaf seeds session dir and/or store with deliberate snapshot vs live mismatch.
2. Sets `req.Enrich` (`""`/`light`/`stale`/`heavy`) and Op list|show.
3. Assert refreshed fields, snapshot retention, Orphaned, and warnings.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.AgentProHome == "" || req.GrokHome == "" {
		t.Fatal("expected AgentProHome and GrokHome from root Setup")
	}
	return nil
}
```
