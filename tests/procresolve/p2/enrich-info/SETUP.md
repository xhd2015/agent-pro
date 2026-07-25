# Scenario

**Feature**: optional Grok title/model enrichment after a hard grok resolve

```
# EnrichInfo gate + injectable LookupGrokInfo
ResolveFromPID(pid, Options{EnrichInfo, LookupGrokInfo, ListProcs, Lsof, GrokHome})
  -> hard grok hit (session from open files)
  -> if EnrichInfo: LookupGrokInfo(home, sessionID) fills GrokTitle/GrokModel
  -> if !EnrichInfo: GrokTitle/GrokModel stay empty (lookup must not be required)
doctest <- Result
```

## Preconditions

- Leaves use a bare-input grok hit (pid 100) with fixture session path so Kind=grok
  and SessionID is known before enrich runs.
- `InjectLookup=true` with fixture title/model so tests never touch real
  `~/.grok` or disk.
- Mode is `enrich` (same Run path as resolve; flag on Options).

## Steps

1. Grouping sets Mode, bare grok fixture procs/open files, and inject lookup
   values.
2. Leaves only flip `EnrichInfo` true vs false.
3. Assert GrokTitle/GrokModel filled or empty accordingly; base hit fields still hold.

## Context

- Lookup is always installed on this branch so **off** proves EnrichInfo gates
  the call (title stays empty even though a lookup would succeed).
- Lookup errors are out of scope for these two leaves (soft warning behavior left
  to implementer; not locked here).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "enrich"
	req.PID = 100
	req.Procs = []FixtureProc{
		{PID: 100, PPID: 1, Cmd: "/usr/local/bin/grok"},
	}
	req.OpenFiles = map[int][]string{
		100: {grokSessionPath(fixtureGrokSessionID)},
	}
	req.InjectLookup = true
	req.LookupTitle = fixtureGrokTitle
	req.LookupModel = fixtureGrokModel
	req.LookupErr = ""
	return nil
}
```
