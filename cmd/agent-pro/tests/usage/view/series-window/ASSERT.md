## Expected

- 200 with one block per requested metric, in the requested order, each naming its
  unit and whether it is drawn as steps.
- `used_percent` holds exactly the point inside the last 24 hours (value 30, the
  run's own recorded timestamp) and drops the 50-hour-old one: the window is applied
  to the records, not to the file list.
- `sessions_total` holds the same single instant with the session count of that
  snapshot, and is marked `step` because a session count accumulates; `used_percent`
  is not.
- `range` reports the server's default range (7d) while the points respect the
  query's `since`.

## Side Effects

- None: the store is read.

## Errors

- None.

```go
import (
"testing"
"time"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
if err != nil {
t.Fatal(err)
}
AssertHTTPStatus(t, resp, 200)
AssertNoWarnings(t, resp)

payload := Series(t, resp)
if payload.Schema != "agent-pro/usage-view/v1" {
t.Fatalf("schema = %q, want agent-pro/usage-view/v1", payload.Schema)
}
if payload.Range != "7d" {
t.Fatalf("range = %q, want the server default 7d", payload.Range)
}
if len(payload.Metrics) != 2 {
t.Fatalf("metrics = %d blocks, want used_percent and sessions_total", len(payload.Metrics))
}
if payload.Metrics[0].Name != "used_percent" || payload.Metrics[1].Name != "sessions_total" {
t.Fatalf("metrics = %s, %s, want the requested order",
payload.Metrics[0].Name, payload.Metrics[1].Name)
}

used := MetricBlock(t, payload, "used_percent")
if used.Unit != "%" || used.Kind != "percent" || used.Step {
t.Fatalf("used_percent block = %+v, want percentages drawn as a line", used)
}
if len(used.Series) != 1 {
t.Fatalf("used_percent series = %+v, want grok only", used.Series)
}
points := SeriesFor(t, used, "grok").Points
if len(points) != 1 {
t.Fatalf("used_percent points = %+v, want only the last 24h", points)
}
if points[0].Value != 30 {
t.Fatalf("used_percent point = %v, want the newest 30", points[0].Value)
}
if delta := time.Since(points[0].TS); delta < 110*time.Minute || delta > 130*time.Minute {
t.Fatalf("used_percent point ts is %s old, want about 2h", delta)
}

sessions := MetricBlock(t, payload, "sessions_total")
if !sessions.Step || sessions.Kind != "counter" {
t.Fatalf("sessions_total block = %+v, want a counter drawn as steps", sessions)
}
sessionPoints := SeriesFor(t, sessions, "grok").Points
if len(sessionPoints) != 1 || sessionPoints[0].Value != 3 {
t.Fatalf("sessions_total points = %+v, want the 3 sessions of the newest snapshot", sessionPoints)
}
if !sessionPoints[0].TS.Equal(points[0].TS) {
t.Fatalf("sessions point %s and usage point %s are different instants", sessionPoints[0].TS, points[0].TS)
}
AssertStoreUnchanged(t, req)
}
```
