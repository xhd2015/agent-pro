# Scenario

**Feature**: resolve Kind=codex from rollout open path under a codex runner

```
# codex candidate + Lsof path …/.codex/sessions/…/rollout-…-<uuid>
ResolveFromPID -> Kind=codex, SessionID=<uuid>, Confidence=hard
```

## Preconditions

- Winning candidate classified as role `codex`.
- Session uuid parsed from `rollout-…-<uuid>` filename under `.codex/sessions/`.

## Steps

1. Leaf installs multi-node tree ending in codex.
2. OpenFiles only on the codex leaf.

## Context

- Fixture uuid: `a1b2c3d4-e5f6-7890-abcd-ef1234567890`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.CodexHome = "/tmp/fake-codex-home"
	return nil
}
```
