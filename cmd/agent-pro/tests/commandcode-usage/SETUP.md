# Scenario

**Feature**: agent-pro commandcode usage against a synthetic Command Code API

```
# harness serves fixtures over httptest, home holds auth.json
test harness -> commandcode.NewClient -> FetchUsageWithOptions -> UsageData

# projection + format turn the merged payload into CLI output
UsageData -> ProjectUsageView -> FormatUsage / FormatUsageJSON
```

## Preconditions

- Package `agent/commandcode` exposes NewClient, FetchUsageWithOptions,
  ProjectUsageView, FormatUsage, and FormatUsageJSON.
- Tests never reach `api.commandcode.ai` or read the real `~/.commandcode`:
  the credential comes from the fixture home, so tests do not mutate process
  environment to shield themselves from `$COMMAND_CODE_API_KEY`.

## Steps

1. Root Setup allocates `req.Home` as `{temp}/.commandcode`, writes auth.json,
   and installs empty query and auth-header recorders.
2. Leaf Setup fills `req.Bodies` (and `req.Statuses` for failure branches).
3. Run serves the fixtures, fetches the usage data, and renders it.

```go
import (
"fmt"
"net/http"
"net/http/httptest"
"net/url"
"os"
"path/filepath"
"strings"
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/commandcode"
"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Home = filepath.Join(t.TempDir(), ".commandcode")
WriteAuth(t, req.Home, "test-key")
req.Queries = map[string]string{}
req.AuthHeaders = map[string]string{}
return nil
}

// WriteAuth writes the credentials file the cmd CLI would have written.
func WriteAuth(t *testing.T, home, key string) {
t.Helper()
if err := os.MkdirAll(home, 0o700); err != nil {
t.Fatalf("mkdir home: %v", err)
}
body := fmt.Sprintf(`{"apiKey":%q,"userName":"tester","keyName":"test"}`, key)
if err := os.WriteFile(filepath.Join(home, commandcode.AuthFileName), []byte(body), 0o600); err != nil {
t.Fatalf("write auth.json: %v", err)
}
}

// ServeFixture starts a fake Command Code API. Each path in bodies answers with
// that body and the status from statuses (200 when unset); unlisted paths
// answer 404 with an error envelope. queries records the raw query string and
// authHeaders the Authorization header per path, so a leaf can assert how the
// fetch scoped and authenticated its requests.
func ServeFixture(t *testing.T, bodies map[string]string, statuses map[string]int, queries map[string]string, authHeaders map[string]string) *httptest.Server {
t.Helper()
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
if queries != nil {
queries[r.URL.Path] = r.URL.RawQuery
}
if authHeaders != nil {
authHeaders[r.URL.Path] = r.Header.Get("Authorization")
}
body, ok := bodies[r.URL.Path]
if !ok {
w.WriteHeader(http.StatusNotFound)
_, _ = w.Write([]byte(`{"success":false,"error":{"code":"NOT_FOUND","status":404,"message":"no fixture for ` + r.URL.Path + `"}}`))
return
}
if status := statuses[r.URL.Path]; status != 0 {
w.WriteHeader(status)
}
_, _ = w.Write([]byte(body))
}))
t.Cleanup(server.Close)
return server
}

// ActivePlanBodies is the payload observed live for an individual-go account
// mid-period: $2.41 left of a 10-credit allowance, 1040 requests this billing
// period, and a weekly window that resets in 7h 4m.
func ActivePlanBodies(now time.Time) map[string]string {
return map[string]string{
commandcode.PathWhoami: `{"success":true,"user":{"id":"u1","userName":"alice","name":"Alice","email":"alice@example.test"}}`,
commandcode.PathCredits: fmt.Sprintf(
`{"credits":{"belowThreshold":false,"creditThreshold":0,"monthlyCredits":2.408591571,"purchasedCredits":0,"freeCredits":0},"windowLimits":{"limited":true,"fiveHour":{"used":0,"cap":3,"exceeded":false,"resetAt":0},"weekly":{"used":0.480063398,"cap":6,"exceeded":false,"resetAt":%d}}}`,
now.Add(7*time.Hour+4*time.Minute).UnixMilli()),
commandcode.PathSubscriptions: fmt.Sprintf(
`{"success":true,"data":{"id":"sub_1","status":"active","planId":"individual-go","quantity":1,"currentPeriodStart":%q,"currentPeriodEnd":%q}}`,
now.Add(-28*24*time.Hour).Format(time.RFC3339), now.Add(5*24*time.Hour).Format(time.RFC3339)),
commandcode.PathUsageSummary: `{"totalCount":1040,"totalCost":7.591408429,"averageCost":0.007299431181730769,"successRate":100,"completedCount":1040,"failedCount":0,"totalTokens":22859644,"totalCredits":7.591408429,"totalMonthlyCredits":7.591408429,"periodBasis":"billing-period"}`,
}
}

// ErrorBody is the API's error envelope for a non-2xx response.
func ErrorBody(code string, status int, message string) string {
return fmt.Sprintf(`{"success":false,"error":{"code":%q,"status":%d,"message":%q}}`, code, status, message)
}

func AssertSuccess(t *testing.T, resp *Response) {
t.Helper()
if resp.Err != nil {
t.Fatalf("operation failed: %v", resp.Err)
}
}

// Lines returns the rendered output split into lines, without the trailing
// newline.
func Lines(output string) []string {
return strings.Split(strings.TrimRight(output, "\n"), "\n")
}

// ParseQuery decodes a recorded raw query string, failing the test when it is
// malformed.
func ParseQuery(t *testing.T, raw string) url.Values {
t.Helper()
values, err := url.ParseQuery(raw)
if err != nil {
t.Fatalf("parse query %q: %v", raw, err)
}
return values
}

// FindLine returns the one rendered line containing substr.
func FindLine(t *testing.T, output, substr string) string {
t.Helper()
for _, line := range Lines(output) {
if strings.Contains(line, substr) {
return line
}
}
t.Fatalf("no line contains %q in:\n%s", substr, output)
return ""
}

// BarParts counts the filled and empty blocks of the first bar on a line.
func BarParts(line string) (filled, empty int) {
for _, r := range line {
switch r {
case '█':
filled++
case '░':
empty++
}
}
return filled, empty
}

// Bar renders an expected progress bar: filled blocks then empty blocks, using
// the same glyphs as the renderer.
func Bar(filled, width int) string {
return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

// AssertLine fails unless the rendered line containing substr equals want.
func AssertLine(t *testing.T, output, substr, want string) {
t.Helper()
if got := FindLine(t, output, substr); got != want {
t.Fatalf("line containing %q\n got: %q\nwant: %q\noutput:\n%s", substr, got, want, output)
}
}

func AssertContains(t *testing.T, output, want string) {
t.Helper()
if !strings.Contains(output, want) {
t.Fatalf("missing %q in:\n%s", want, output)
}
}

func AssertAbsent(t *testing.T, output, unwanted string) {
t.Helper()
if strings.Contains(output, unwanted) {
t.Fatalf("unexpected %q in:\n%s", unwanted, output)
}
}
```
