# Scenario

**Feature**: failed provider fetches warn on stderr and still get a snapshot

```
fixtures removed -> grok + codex fetches fail, Command Code still answers
usage collect
stderr: warning: grok: ...   warning: codex: ...
stdout: table with grok/codex "(failed)", then "3 snapshots appended under ..."
        exit 0
```

## Preconditions

- `AGENT_PRO_USAGE_FIXTURE_DIR` points at an empty directory, so the grok and
  codex fetches fail locally instead of reaching the network.
- Command Code keeps its credentials and fake API, so the run is a partial
  failure: the case where one provider is down and the others must still be
  recorded.

## Steps

1. Point the fixture directory at the empty one; leave the output human-readable.

```go
import (
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.FixtureDir = req.EmptyFixtureDir
return nil
}
```
