# Scenario

**Feature**: `usage.Collect` reads every provider's usage over its API and counts that
provider's on-disk sessions, producing one snapshot record per provider

```
# fixed clock, isolated homes, fixture-transport grok + codex, fake Command Code API
collectNow -> CollectOptions{Now, GrokHome, CodexHome, CommandCodeHome, FixtureDir}
Collect(ctx, opts) -> []Record{grok, codex, commandcode}   # parallel, order preserved

# per provider
fetch API    -> UsageBlock{ok, endpoint, values, display, error}
walk <home>  -> SessionsBlock{total, oldest, newest}
```

## Preconditions

- Grok and codex fetches are served by JSON fixtures (`CollectOptions.FixtureDir`),
  never by the network; the Command Code API is an `httptest` server.
- Leaf `providers/usage-failure-keeps-sessions` runs without fixtures and without
  Command Code credentials, so every fetch fails locally instead of hanging.
- Every home is a fresh temp dir; a leaf seeds only the session tree it is about.
- Nothing here touches `~/.agent-pro`, `~/.grok`, `~/.codex` or `~/.commandcode`.

## Steps

1. Root `Setup` allocates the clock, the four homes (three with credentials, one
   bare), the default fixtures, and the default Command Code API payload.
2. Group `Setup` documents the branch; leaf `Setup` seeds session trees and, where
   the leaf is about a fallback or a failure, removes a fixture or credential.
3. `Run` materializes the fixtures, starts the fake API when bodies are given, and
   calls `usage.Collect`.
4. Leaf `Assert` checks the usage block, the session block, and the record order.

## Context

- Snapshot timestamps are UTC truncated to the second, so `ts` is exactly `req.Now`.
- The fixture file names are fixed by `agent/usage`: `grok-billing.json`,
  `grok-billing-credits.json`, `codex-usage.json`, `codex-wham.json`.
- Command Code plan `individual-go` means a 10 credit monthly allowance named `Go`.

```go
import (
"fmt"
"os"
"path/filepath"
"strings"
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
"github.com/xhd2015/doctest/session"
)

// collectNow is the instant every leaf collects at.
var collectNow = time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)

// grokAuth and codexAuth are the credential files the provider fetches read.
const (
grokAuth = `{"https://accounts.x.ai::client":{"key":"fixture-access-token","auth_mode":"oidc",` +
`"email":"grok@example.com","expires_at":"2030-01-01T00:00:00Z"}}`
codexAuth = `{"auth_mode":"chatgpt","last_refresh":"2026-09-15T00:00:00Z",` +
`"tokens":{"access_token":"fixture-access-token","account_id":"00000000-0000-4000-8000-000000000001"}}`
commandCodeAuth = `{"apiKey":"fixture-key","userName":"tester","keyName":"test"}`
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Now = collectNow
tmp := t.TempDir()
req.GrokHome = filepath.Join(tmp, "grok")
req.CodexHome = filepath.Join(tmp, "codex")
req.CommandCodeHome = filepath.Join(tmp, "commandcode")
req.BareCommandCodeHome = filepath.Join(tmp, "commandcode-nocreds")

for _, dir := range []string{req.GrokHome, req.CodexHome, req.CommandCodeHome, req.BareCommandCodeHome} {
if err := os.MkdirAll(dir, 0o755); err != nil {
return fmt.Errorf("mkdir %s: %w", dir, err)
}
}
credentials := map[string]string{
filepath.Join(req.GrokHome, "auth.json"):        grokAuth,
filepath.Join(req.CodexHome, "auth.json"):       codexAuth,
filepath.Join(req.CommandCodeHome, "auth.json"): commandCodeAuth,
}
for path, body := range credentials {
if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
return fmt.Errorf("write %s: %w", path, err)
}
}

req.CommandCodeCredentialed = true
req.Fixtures = DefaultFixtures()
req.CommandCodeBodies = commandCodeBodies(req.Now)
return nil
}

// DefaultFixtures are the grok and codex payloads: a monthly grok allowance
// 73% used, a weekly credit period 2% used, and 62% of the codex limit used.
func DefaultFixtures() map[string]string {
return map[string]string{
"grok-billing.json": `{"config":{"monthlyLimit":{"val":100},"used":{"val":73},` +
`"onDemandCap":{"val":0},"billingPeriodStart":"2026-08-01T00:00:00+00:00",` +
`"billingPeriodEnd":"2026-09-01T00:00:00+00:00","history":[]}}`,
"grok-billing-credits.json": `{"config":{"billingPeriodStart":"2026-08-28T00:55:25.179446+00:00",` +
`"billingPeriodEnd":"2026-09-04T00:55:25.179446+00:00","creditUsagePercent":2.0,` +
`"currentPeriod":{"start":"2026-08-28T00:55:25.179446+00:00",` +
`"end":"2026-09-04T00:55:25.179446+00:00","type":"USAGE_PERIOD_TYPE_WEEKLY"}}}`,
"codex-usage.json": `{"plan_type":"business","email":"codex@example.com","rate_limit":null,` +
`"spend_control":{"individual_limit":{"used_percent":62,"remaining_percent":38,` +
`"reset_at":1788220800}}}`,
"codex-wham.json": `{"plan_type":"business","email":"codex@example.com","rate_limit":null,` +
`"spend_control":{"individual_limit":{"used_percent":62,"remaining_percent":38,` +
`"reset_at":1788220800}}}`,
}
}

// WriteGrokSession writes one grok session summary under <home>/sessions/<rel>.
// rel may be nested ("parent/subagents/child") so leaves can build child sessions.
func WriteGrokSession(t *testing.T, home, rel, createdAt, updatedAt string) string {
t.Helper()
dir := filepath.Join(home, "sessions", rel)
if err := os.MkdirAll(dir, 0o755); err != nil {
t.Fatalf("mkdir grok session: %v", err)
}
path := filepath.Join(dir, "summary.json")
body := fmt.Sprintf(`{"created_at":%q,"updated_at":%q,"session_summary":"test"}`, createdAt, updatedAt)
if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
t.Fatalf("write grok summary: %v", err)
}
return path
}

// WriteCodexRollout writes a codex rollout transcript. The file name carries
// the session start (local time) and modTime stands in for the last activity.
func WriteCodexRollout(t *testing.T, home, day, stamp, id string, modTime time.Time) string {
t.Helper()
dir := filepath.Join(home, "sessions", day)
if err := os.MkdirAll(dir, 0o755); err != nil {
t.Fatalf("mkdir codex session: %v", err)
}
path := filepath.Join(dir, "rollout-"+stamp+"-"+id+".jsonl")
if err := os.WriteFile(path, []byte(`{"type":"session_meta"}`+"\n"), 0o644); err != nil {
t.Fatalf("write codex rollout: %v", err)
}
if err := os.Chtimes(path, modTime, modTime); err != nil {
t.Fatalf("set codex rollout mtime: %v", err)
}
return path
}

// WriteCommandCodeTranscript writes one Command Code transcript under
// <home>/projects/<project>/, with firstRecord as its first JSON line.
func WriteCommandCodeTranscript(t *testing.T, home, project, name, firstRecord string, modTime time.Time) string {
t.Helper()
dir := filepath.Join(home, "projects", project)
if err := os.MkdirAll(dir, 0o755); err != nil {
t.Fatalf("mkdir commandcode project: %v", err)
}
path := filepath.Join(dir, name)
if err := os.WriteFile(path, []byte(firstRecord+"\n"), 0o644); err != nil {
t.Fatalf("write transcript: %v", err)
}
if err := os.Chtimes(path, modTime, modTime); err != nil {
t.Fatalf("set transcript mtime: %v", err)
}
return path
}

// AssertUsageOK fails unless the provider's usage fetch succeeded.
func AssertUsageOK(t *testing.T, rec usage.Record) {
t.Helper()
if !rec.Usage.OK {
t.Fatalf("%s usage not ok: %s", rec.Provider, rec.Usage.Error)
}
}

// AssertUsageFailed fails unless the provider's usage fetch failed with an
// error containing want.
func AssertUsageFailed(t *testing.T, rec usage.Record, want string) {
t.Helper()
if rec.Usage.OK {
t.Fatalf("%s usage unexpectedly ok: %v", rec.Provider, rec.Usage.Values)
}
if !strings.Contains(rec.Usage.Error, want) {
t.Fatalf("%s usage error = %q, want it to contain %q", rec.Provider, rec.Usage.Error, want)
}
}

// Value returns one numeric usage value, failing when the key is absent.
func Value(t *testing.T, rec usage.Record, key string) float64 {
t.Helper()
value, ok := rec.Usage.Values[key]
if !ok {
t.Fatalf("%s has no usage value %q (have %v)", rec.Provider, key, rec.Usage.Values)
}
return value
}

// Display returns one display string, failing when the key is absent.
func Display(t *testing.T, rec usage.Record, key string) string {
t.Helper()
value, ok := rec.Usage.Display[key]
if !ok {
t.Fatalf("%s has no display value %q (have %v)", rec.Provider, key, rec.Usage.Display)
}
return value
}

// AssertSessions checks the session block. oldest and newest are RFC3339 UTC
// strings, or empty for "not reported".
func AssertSessions(t *testing.T, rec usage.Record, total int, oldest, newest string) {
t.Helper()
if rec.Sessions.Error != "" {
t.Fatalf("%s session walk failed: %s", rec.Provider, rec.Sessions.Error)
}
if rec.Sessions.Total != total {
t.Fatalf("%s sessions total = %d, want %d", rec.Provider, rec.Sessions.Total, total)
}
AssertSessionTime(t, rec.Provider, "oldest", rec.Sessions.Oldest, oldest)
AssertSessionTime(t, rec.Provider, "newest", rec.Sessions.Newest, newest)
}

func AssertSessionTime(t *testing.T, provider usage.ProviderID, field string, got *time.Time, want string) {
t.Helper()
if want == "" {
if got != nil {
t.Fatalf("%s sessions %s = %s, want unset", provider, field, got.UTC().Format(time.RFC3339))
}
return
}
if got == nil {
t.Fatalf("%s sessions %s is unset, want %s", provider, field, want)
}
if formatted := got.UTC().Format(time.RFC3339); formatted != want {
t.Fatalf("%s sessions %s = %s, want %s", provider, field, formatted, want)
}
}

// AssertValues compares a whole value set, reporting every difference at once.
func AssertValues(t *testing.T, rec usage.Record, want map[string]float64) {
t.Helper()
for key, value := range want {
if got := Value(t, rec, key); got != value {
t.Fatalf("%s usage %s = %v, want %v", rec.Provider, key, got, value)
}
}
for key := range rec.Usage.Values {
if _, ok := want[key]; !ok {
t.Fatalf("%s has unexpected usage value %q = %v", rec.Provider, key, rec.Usage.Values[key])
}
}
}
```
