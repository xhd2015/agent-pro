# usage api collect tests

Doc-style tests for the HTTP collect path of
`github.com/xhd2015/agent-pro/agent/usage`: `Collect` reads every provider's
account usage over its API and counts the sessions that provider has on disk,
in parallel, producing one `Record` per provider.

# DSN (Domain Specific Notion)

`agent-pro usage collect` appends one snapshot per provider under
`$AGENT_PRO_HOME/usages/<provider>/<date>/<time>-snapshot.jsonl`. This tree
covers what goes into such a snapshot: the API usage block and the on-disk
session block.

**Participants**

- **Collect** — `usage.Collect(ctx, CollectOptions)`; one goroutine per
  provider, always the same three (`grok`, `codex`, `commandcode`) in that
  order.
- **grok** — billing API through `shell/grok/usage`: `/v1/billing` for the
  monthly allowance, the same URL with `?format=credits` for the weekly
  credit period.
- **codex** — `backend-api/codex/usage`, falling back to
  `backend-api/wham/usage`.
- **commandcode** — `agent/commandcode` composite fetch: whoami scopes
  credits and subscription, the subscription period bounds
  `/alpha/usage/summary`.
- **Session trees** — `<home>/sessions` for grok and codex, `<home>/projects`
  for commandcode. Counting only reads local files, so it still runs when a
  usage fetch fails.
- **Fixture transport** — `CollectOptions.FixtureDir`
  (`AGENT_PRO_USAGE_FIXTURE_DIR` in the CLI) replaces the grok and codex HTTP
  call with a JSON file read; the Command Code API is served by `httptest`.

```
collect cycle
  grok        -> fixture grok-billing.json        -> UsageBlock{ok, values, display}
  codex       -> fixture codex-usage.json         -> UsageBlock{ok, values, display}
  commandcode -> httptest whoami/credits/sub/summary -> UsageBlock{ok, values, display}
  each provider, in parallel:
    sessions  -> walk <home> tree                 -> SessionsBlock{total, oldest, newest}
doctest <- []Record{grok, codex, commandcode}
```

## Version

0.0.1

## Decision Tree

```
agent/usage/tests/collect/
├── DOCTEST.md
├── SETUP.md                                        # homes, auth, fixtures, fake Command Code API
├── providers/                                      # one Collect cycle over all three providers
│   ├── all-ok/                                     # billing + codex usage + plan → 3 ok records
│   ├── grok-weekly-credits/                        # billing fixture absent → credits → weekly
│   ├── codex-wham-fallback/                        # codex/usage fixture absent → wham/usage
│   ├── commandcode-plan-credits/                   # credits, window limits, summary → values
│   └── usage-failure-keeps-sessions/               # every fetch fails; sessions still counted
└── sessions/                                       # on-disk session shapes per provider
    ├── grok-includes-children/                     # subagent + fork children count; malformed skipped
    ├── codex-rollout-names/                        # oldest from the rollout name, newest from mtime
    ├── commandcode-transcripts/                    # checkpoints sidecar excluded; createdAt fallback
    └── missing-home/                               # provider never used → total 0, no error
```

(Store layout, append and read back are a sibling doctest root: `../../store`.)

Parameter ranking (most → least significant):

1. **Fetch outcome** — usage ok vs failed (a failed record still carries sessions)
2. **Provider** — grok / codex / commandcode payload shapes and endpoints
3. **Endpoint fallback** — billing → credits, codex/usage → wham/usage
4. **Session tree shape** — which files count, and where their timestamps come from
5. **Session block fields** — total, oldest, newest

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `providers/all-ok` | All three providers answer: record order, usage values, session totals |
| 2 | `providers/grok-weekly-credits` | Billing fixture absent: grok reports the weekly credit period |
| 3 | `providers/codex-wham-fallback` | Codex usage fixture absent: the wham endpoint still reports usage |
| 4 | `providers/commandcode-plan-credits` | Credits, window limits and summary become values plus display strings |
| 5 | `providers/usage-failure-keeps-sessions` | No fixtures, no credentials: `ok:false` with error, sessions unaffected |
| 6 | `sessions/grok-includes-children` | Main, subagent and fork sessions count; a malformed summary is skipped |
| 7 | `sessions/codex-rollout-names` | Oldest from the rollout file name, newest from the file modification time |
| 8 | `sessions/commandcode-transcripts` | Transcripts count, `.checkpoints.jsonl` sidecars do not |
| 9 | `sessions/missing-home` | A provider with no home reports total 0 and no timestamps |

## How to Run

```sh
doctest vet ./agent/usage/tests/collect
doctest test -v ./agent/usage/tests/collect
doctest test -v ./agent/usage/tests/collect/providers/all-ok
doctest test -v ./agent/usage/tests/collect/sessions/grok-includes-children
```

```go
import (
"context"
"net/http"
"net/http/httptest"
"os"
"path/filepath"
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/commandcode"
"github.com/xhd2015/agent-pro/agent/usage"
"github.com/xhd2015/doctest/session"
)

type Request struct {
// Now is the injected collect clock; every leaf runs at this fixed instant.
Now time.Time
// GrokHome and CodexHome hold auth.json plus their session trees.
GrokHome string
CodexHome string
// CommandCodeHome holds auth.json and the project transcripts; the fake API
// describes it when CommandCodeBodies is set.
CommandCodeHome string
// BareCommandCodeHome has no credentials, so a collect that uses it fails
// locally instead of reaching the real API.
BareCommandCodeHome string
// CommandCodeCredentialed selects which Command Code home the collect uses.
CommandCodeCredentialed bool
// Fixtures maps a fixture file name to the JSON body written for it. A
// non-nil empty map makes the grok and codex fetches fail without network.
Fixtures map[string]string
// CommandCodeBodies maps a Command Code API path to its response body; nil
// means no fake API is served.
CommandCodeBodies map[string]string
}

type Response struct {
Records    []usage.Record
ByProvider map[usage.ProviderID]usage.Record
}

// Record returns one provider's record, failing when it is missing.
func (r *Response) Record(t *testing.T, id usage.ProviderID) usage.Record {
t.Helper()
rec, ok := r.ByProvider[id]
if !ok {
t.Fatalf("no record collected for provider %q", id)
}
return rec
}

// Providers lists the collected providers in record order.
func (r *Response) Providers() []string {
out := make([]string, 0, len(r.Records))
for _, rec := range r.Records {
out = append(out, string(rec.Provider))
}
return out
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
t.Helper()
home := req.CommandCodeHome
if !req.CommandCodeCredentialed {
home = req.BareCommandCodeHome
}
opts := usage.CollectOptions{
Now:             func() time.Time { return req.Now },
GrokHome:        req.GrokHome,
CodexHome:       req.CodexHome,
CommandCodeHome: home,
}
if req.Fixtures != nil {
opts.FixtureDir = WriteFixtures(t, req.Fixtures)
}
if req.CommandCodeBodies != nil {
opts.CommandCodeAPIURL = ServeCommandCode(t, req.CommandCodeBodies).URL
}

records := usage.Collect(context.Background(), opts)
resp := &Response{Records: records, ByProvider: map[usage.ProviderID]usage.Record{}}
for _, rec := range records {
resp.ByProvider[rec.Provider] = rec
}
return resp, nil
}

// ServeCommandCode starts a fake Command Code API: every path in bodies answers
// that body, and any other path answers 404 with the API's error envelope.
func ServeCommandCode(t *testing.T, bodies map[string]string) *httptest.Server {
t.Helper()
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
body, ok := bodies[r.URL.Path]
if !ok {
w.WriteHeader(http.StatusNotFound)
_, _ = w.Write([]byte(`{"success":false,"error":{"code":"NOT_FOUND","status":404,"message":"no fixture"}}`))
return
}
_, _ = w.Write([]byte(body))
}))
t.Cleanup(server.Close)
return server
}

// WriteFixtures materializes the grok/codex fixture bodies in a fresh
// directory and returns it as the collect's replacement for the HTTP call.
func WriteFixtures(t *testing.T, fixtures map[string]string) string {
t.Helper()
dir := t.TempDir()
for name, body := range fixtures {
if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
t.Fatalf("write fixture %s: %v", name, err)
}
}
return dir
}

// commandCodeBodies describes a Command Code account mid-period: 7.5 of a 10
// credit monthly allowance left, 2.5 credits spent over 120 requests, and
// window limits half used (5-hour) and a quarter used (weekly).
func commandCodeBodies(now time.Time) map[string]string {
periodStart := now.Add(-7 * 24 * time.Hour).Format(time.RFC3339)
periodEnd := now.Add(3 * 24 * time.Hour).Format(time.RFC3339)
return map[string]string{
commandcode.PathWhoami: `{"success":true,"user":{"id":"u1","userName":"alice","name":"Alice","email":"alice@example.test"}}`,
commandcode.PathCredits: `{"credits":{"belowThreshold":false,"creditThreshold":0,"monthlyCredits":7.5,` +
`"purchasedCredits":0,"freeCredits":0},"windowLimits":{"limited":true,` +
`"fiveHour":{"used":1.5,"cap":3,"exceeded":false,"resetAt":0},` +
`"weekly":{"used":1.5,"cap":6,"exceeded":false,"resetAt":1789600000000}}}`,
commandcode.PathSubscriptions: `{"success":true,"data":{"id":"sub_1","status":"active","planId":"individual-go",` +
`"quantity":1,"currentPeriodStart":"` + periodStart + `","currentPeriodEnd":"` + periodEnd + `"}}`,
commandcode.PathUsageSummary: `{"totalCount":120,"totalCost":2.5,"averageCost":0.020833333,` +
`"successRate":96,"completedCount":115,"failedCount":5,"totalTokens":1000,"totalTokensIn":600,` +
`"totalTokensOut":400,"totalCredits":2.5,"totalMonthlyCredits":2.5,"periodBasis":"billing-period"}`,
}
}
```
