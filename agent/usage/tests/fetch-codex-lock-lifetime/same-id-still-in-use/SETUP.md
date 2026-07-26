# Scenario

**Feature**: after early flock release, concurrent reserve of the **same** custom session id
while the first usage fetch session is still live must fail with “already in use”
(not registry lock-busy)

```
FetchStatus live (claim or registry for codex-status-usage)
peer ReserveCustomSessionID(codex-status-usage) -> error "already in use"
# must NOT be "registry lock busy" / flock timeout
```

## Preconditions

- Inherits Codex blocking fake and Mode=lock-during-fetch.
- SameIDProbe=true so Run records concurrent same-id reserve error.

## Steps

1. Enable SameIDProbe.
2. Mid-fetch, after claim/registry visible, attempt ReserveCustomSessionID(same id).
3. Assert error mentions already in use; assert flock itself was free (LockAcquiredDuring).

## Context

- With deferred whole-fetch release, same-id reserve often fails with lock-busy first —
  wrong failure mode and blocks all peers.
- Early release restores “already in use” semantics while allowing other session ids.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SameIDProbe = true
	return nil
}
```
