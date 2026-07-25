# Scenario

**Feature**: resolve Kind=grok from open files under a grok runner

```
# grok candidate + Lsof path …/.grok/sessions/…/<uuid>/…
ResolveFromPID -> Kind=grok, SessionID=<uuid>, Confidence=hard
```

## Preconditions

- Winning candidate is classified as role `grok`.
- Session uuid parsed from path segment under `.grok/sessions/`.

## Steps

1. Leaf chooses bare-input vs agent-run multi-level topology.
2. Install grok session open path via `OpenFiles`.

## Context

- Fixture uuid: `019fabcdef-1234-5678-9abc-def012345678`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Ensure grok home is set for this branch (paths still absolute in Lsof).
	req.GrokHome = "/tmp/fake-grok-home"
	return nil
}
```
