# Scenario

**Feature**: nearest grok ancestor yields a hard session id from open files

```
FindAncestorGrok(start) -> grok proc
ResolveFromAncestors(start) -> Kind=grok, SessionID from that grok's Lsof
```

## Preconditions

- A real grok runner sits on the start-then-PPID chain (not `grok update`).
- That grok’s `Lsof` contains `/.grok/sessions/…/<uuid>/…`.

## Steps

1. Leaf installs the PPID chain and start pid.
2. Open files on the winning grok only (or on both nested groks).
3. Assert ancestor pid + hard session id.

## Context

- Default fixture uuid: `019fabcdef-1234-5678-9abc-def012345678`.
- Cmdline `--resume` / `--session-id` on the grok must not win.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.GrokHome = "/tmp/fake-grok-home"
	return nil
}
```
