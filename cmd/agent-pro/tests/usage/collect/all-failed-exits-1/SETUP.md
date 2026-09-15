# Scenario

**Feature**: a run where no provider answers exits non-zero

```
fixtures removed + Command Code home without credentials
usage collect --json
stderr: 3 warnings, then "agent-pro: no usage collected: every provider failed"
        exit 1; stdout still carries the three ok:false records, and the
        three snapshot files exist
```

## Preconditions

- The fixture directory is empty, so grok and codex fail before any HTTP call.
- `ServeCommandCode` is off and the Command Code home has no `auth.json`, so that
  fetch fails locally too.
- A cron job or a monitoring script needs this to be a *failure*: exit 0 would
  hide that no usage is being recorded.

## Steps

1. Empty the fixture directory, drop the fake Command Code API and use the
   credential-less home.

```go
import (
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.FixtureDir = req.EmptyFixtureDir
req.ServeCommandCode = false
req.CommandCodeHome = req.BareCommandCodeHome
req.Args = []string{"--json"}
return nil
}
```
