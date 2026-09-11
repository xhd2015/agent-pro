## Expected

- Top-level keys are exactly `whoami`, `credits`, `subscription`, `summary`,
  and `errors`: the shape the cmd CLI's own fetch returns, so a consumer of
  `usage --json` can address each endpoint payload by name.
- `errors` is an empty array rather than `null`, so a consumer can range over
  it without a nil check.
- Endpoint payloads stay at API precision: `2.408591571` is not rounded to the
  overlay's two decimals.
- Output is indented with two spaces, matching the CLI's JSON printer.

## Errors

- None.

```go
import (
"encoding/json"
"strings"
"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
AssertSuccess(t, resp)

var payload map[string]json.RawMessage
if err := json.Unmarshal(resp.JSON, &payload); err != nil {
t.Fatalf("unmarshal payload: %v\n%s", err, resp.JSON)
}

wantKeys := []string{"whoami", "credits", "subscription", "summary", "errors"}
if len(payload) != len(wantKeys) {
t.Fatalf("payload has %d keys, want %d:\n%s", len(payload), len(wantKeys), resp.JSON)
}
for _, key := range wantKeys {
if _, ok := payload[key]; !ok {
t.Fatalf("missing key %q in:\n%s", key, resp.JSON)
}
}

var warnings []string
if err := json.Unmarshal(payload["errors"], &warnings); err != nil {
t.Fatalf("errors is not a string array: %v (%s)", err, payload["errors"])
}
if len(warnings) != 0 {
t.Fatalf("errors=%v, want empty", warnings)
}

var credits struct {
Credits struct {
MonthlyCredits float64 `json:"monthlyCredits"`
} `json:"credits"`
}
if err := json.Unmarshal(payload["credits"], &credits); err != nil {
t.Fatalf("decode credits: %v", err)
}
if credits.Credits.MonthlyCredits != 2.408591571 {
t.Fatalf("monthlyCredits=%v, want 2.408591571 unrounded", credits.Credits.MonthlyCredits)
}

var subscription struct {
PlanID             string `json:"planId"`
CurrentPeriodStart string `json:"currentPeriodStart"`
}
if err := json.Unmarshal(payload["subscription"], &subscription); err != nil {
t.Fatalf("decode subscription: %v", err)
}
if subscription.PlanID != "individual-go" {
t.Fatalf("planId=%q, want individual-go", subscription.PlanID)
}
if subscription.CurrentPeriodStart != resp.Data.Subscription.CurrentPeriodStart {
t.Fatalf("currentPeriodStart=%q, want %q", subscription.CurrentPeriodStart, resp.Data.Subscription.CurrentPeriodStart)
}

var summary struct {
TotalCount int `json:"totalCount"`
}
if err := json.Unmarshal(payload["summary"], &summary); err != nil {
t.Fatalf("decode summary: %v", err)
}
if summary.TotalCount != 1040 {
t.Fatalf("totalCount=%d, want 1040", summary.TotalCount)
}

if !strings.Contains(string(resp.JSON), "\n  \"whoami\"") {
t.Fatalf("JSON is not indented with two spaces:\n%s", resp.JSON)
}
}
```
