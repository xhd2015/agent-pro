# Scenario

**Feature**: response `counts` present; independent of `q`; `total` respects `q`

```
seed 5 (running=1, done=finished+idle=4)
GET /api/agent-run/sessions?limit=30&offset=0
  -> counts.all=5, counts.running=1, counts.done=4
GET ?q=UNIQUE-QUERY-TOKEN&limit=30
  -> total=1 (q applied)
  -> counts still all=5, running=1, done=4 (q ignored for counts)
```

## Preconditions

- `done` = finished + idle (not only literal status `done`).
- Expect RED until counts object is returned.

## Steps

1. Seed five; start web.
2. GET unfiltered with limit (forces pagination envelope).
3. GET with `q` that matches only delta.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Scenario = "counts-present"
	if err := seedSessions(t, req, defaultFiveSessions()); err != nil {
		return err
	}
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}
	req.HTTPSteps = []HTTPStep{
		{Name: "all", Method: "GET", Path: sessionsPath("limit=30&offset=0")},
		{Name: "with-q", Method: "GET", Path: sessionsPath("q=UNIQUE-QUERY-TOKEN&limit=30&offset=0")},
	}
	return nil
}
```
