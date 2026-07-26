# Scenario

**Feature**: `limit` + `offset` return non-overlapping newest-first pages with total/has_more

```
seed 5 metas
GET ?limit=2&offset=0 -> [epsilon, delta]; total=5; has_more=true
GET ?limit=2&offset=2 -> [gamma, beta]; no overlap
GET ?limit=2&offset=4 -> [alpha]; has_more=false
```

## Preconditions

- Five seeded sessions.
- Expect RED until limit/offset pagination is implemented (today returns full list).

## Steps

1. Seed five; start web.
2. Three GET steps with limit=2 and offsets 0, 2, 4.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Scenario = "limit-offset-page"
	if err := seedSessions(t, req, defaultFiveSessions()); err != nil {
		return err
	}
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}
	req.HTTPSteps = []HTTPStep{
		{Name: "page0", Method: "GET", Path: sessionsPath("limit=2&offset=0")},
		{Name: "page1", Method: "GET", Path: sessionsPath("limit=2&offset=2")},
		{Name: "page2", Method: "GET", Path: sessionsPath("limit=2&offset=4")},
	}
	return nil
}
```
