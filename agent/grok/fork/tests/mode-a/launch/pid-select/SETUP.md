# Scenario

**Feature**: `--pid` selects which process to walk from

```
# default start 6000 → grok 4242 (fixture session)
# alt grok 7000 (alt session), not on that chain
fork.Main(["--pid", "7000"])
  -> follow-up --session-id <alt>
```

## Steps

1. Add alt grok pid 7000 with its own session + Lsof.
2. Args `["--pid", "7000"]`. Injected `req.PID` stays 6000 (must be overridden).

```go
import (
	"strconv"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	altDir := seedSession(t, req.GrokHome, fixtureAltSessionID, req.Workspace)
	req.Procs = append(req.Procs, FixtureProc{
		PID:  pidAltGrok,
		PPID: 1,
		Cmd:  "/usr/local/bin/grok",
	})
	req.OpenFiles[pidAltGrok] = []string{lsofPath(altDir)}
	req.Args = []string{"--pid", strconv.Itoa(pidAltGrok)}
	return nil
}
```
