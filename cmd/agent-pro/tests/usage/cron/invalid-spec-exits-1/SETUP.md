# Scenario

**Feature**: a `--cron` value that is not a schedule fails before collecting anything

```
usage collect --cron 'every 5 min'
stderr: agent-pro: invalid --cron spec "every 5 min": expected 5 fields ...; exit 1
```

## Preconditions

- `every 5 min` looks like a schedule but is not one: the accepted forms are a
  5-field cron, `@hourly|@daily|@midnight|@weekly|@monthly`, or `@every <duration>`.
- The spec is validated before the first cycle, so a typo never writes a single
  snapshot.

## Steps

1. Pass the invalid spec.

```go
import (
"testing"

"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Args = []string{"--cron", "every 5 min"}
return nil
}
```
