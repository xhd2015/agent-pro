# Scenario

**Feature**: only `grok update` in the tree is not a session runner

```
pid 500 "grok update" (or /usr/local/bin/grok update)
  Lsof may even look session-like; still Kind=none
ResolveFromPID(500) -> Kind=none (excluded)
```

## Preconditions

- Cmd is classified as update utility, not interactive/session grok.
- Even if open files existed, update must not be treated as a candidate; this
  leaf uses empty Lsof to lock the exclusion on classification.

## Steps

1. Set `PID=500` with Cmd containing `grok` and `update` as the subcommand.
2. Empty OpenFiles.

## Context

- Exclusion is by argv classification (`grok update`), not by missing files alone
  (see also `plain-bash` for empty-Lsof without grok).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.PID = 500
	req.Procs = []FixtureProc{
		{PID: 500, PPID: 1, Cmd: "/usr/local/bin/grok update"},
	}
	req.OpenFiles = map[int][]string{}
	return nil
}
```
