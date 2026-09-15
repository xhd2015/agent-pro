# Scenario

**Feature**: the `agent-pro usage` CLI against temp homes, fixture usage and a fake Command Code API

```
go build -C cmd -o <tmp>/bin/agent-pro ./agent-pro
env AGENT_PRO_HOME=<tmp>/agent-pro GROK_HOME=... CODEX_HOME=... \
    AGENT_PRO_USAGE_FIXTURE_DIR=<tmp>/fixtures COMMANDCODE_*=<fake API>
agent-pro usage collect [--json] [--cron SPEC] ...
doctest <- stdout, stderr, exit code, files under <tmp>/agent-pro/usages
```

## Preconditions

- The binary is built from this checkout, so CLI flag parsing, exit codes and the
  stdout/stderr split are all covered.
- Every home is a temp dir: the parent process's `AGENT_PRO_HOME`, `GROK_HOME`,
  `CODEX_HOME`, `COMMAND_CODE_API_KEY` and `COMMANDCODE_*` are stripped from the
  child environment, so a developer's real accounts can never be read or charged.
- grok and codex usage comes from fixture files; Command Code comes from an
  `httptest` server unless a leaf deliberately turns it off.

## Steps

1. Root `Setup` builds the binary, creates the homes, credentials, fixtures and an
   empty fixture directory, and seeds one session per provider.
2. Group `Setup` sets the command shape (`WithHomes` for collect).
3. Leaf `Setup` sets the arguments and, for failure leaves, the empty fixture
   directory or the credential-less Command Code home.
4. `Run` executes the binary, then reads the store back.

## Context

- `usage collect` timestamps records with the real clock (UTC, whole seconds), so
  assertions treat `ts` as "now", not as a fixture.
- Snapshot files are named with local time; the day directory is therefore today's
  local date.
- Exit status is 0 while at least one provider's usage was fetched, 1 when every
  provider failed or `--cron` is invalid.

```go
import (
"encoding/json"
"fmt"
"io"
"io/fs"
"net/http"
"os"
"os/exec"
"path/filepath"
"runtime"
"strings"
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/commandcode"
"github.com/xhd2015/agent-pro/agent/usage"
"github.com/xhd2015/doctest/session"
)

const (
grokAuth = `{"https://accounts.x.ai::client":{"key":"fixture-access-token","auth_mode":"oidc",` +
`"email":"grok@example.com","expires_at":"2030-01-01T00:00:00Z"}}`
codexAuth = `{"auth_mode":"chatgpt","last_refresh":"2026-09-15T00:00:00Z",` +
`"tokens":{"access_token":"fixture-access-token","account_id":"00000000-0000-4000-8000-000000000001"}}`
commandCodeAuth = `{"apiKey":"fixture-key","userName":"tester","keyName":"test"}`
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
repoRoot := filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "../../../.."))
if _, err := os.Stat(filepath.Join(repoRoot, "go.mod")); err != nil {
return fmt.Errorf("repo root not found: %w", err)
}

tmp := t.TempDir()
binDir := filepath.Join(tmp, "bin")
if err := os.MkdirAll(binDir, 0o755); err != nil {
return err
}
req.Bin = filepath.Join(binDir, "agent-pro")
build := exec.Command(runtime.GOROOT()+"/bin/go", "build", "-C", "cmd", "-o", req.Bin, "./agent-pro")
build.Dir = repoRoot
if out, err := build.CombinedOutput(); err != nil {
return fmt.Errorf("build agent-pro: %w\n%s", err, string(out))
}

req.Home = filepath.Join(tmp, "agent-pro")
req.GrokHome = filepath.Join(tmp, "grok")
req.CodexHome = filepath.Join(tmp, "codex")
req.CommandCodeHome = filepath.Join(tmp, "commandcode")
req.BareCommandCodeHome = filepath.Join(tmp, "commandcode-nocreds")
req.FixtureDir = filepath.Join(tmp, "fixtures")
req.EmptyFixtureDir = filepath.Join(tmp, "no-fixtures")

for _, dir := range []string{
req.Home, req.GrokHome, req.CodexHome, req.CommandCodeHome,
req.BareCommandCodeHome, req.FixtureDir, req.EmptyFixtureDir,
} {
if err := os.MkdirAll(dir, 0o755); err != nil {
return err
}
}
credentials := map[string]string{
filepath.Join(req.GrokHome, "auth.json"):        grokAuth,
filepath.Join(req.CodexHome, "auth.json"):       codexAuth,
filepath.Join(req.CommandCodeHome, "auth.json"): commandCodeAuth,
}
for path, body := range credentials {
if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
return err
}
}
for name, body := range defaultFixtures() {
if err := os.WriteFile(filepath.Join(req.FixtureDir, name), []byte(body), 0o644); err != nil {
return err
}
}

SeedSessions(t, req)
req.ServeCommandCode = true
req.Timeout = 60 * time.Second
return nil
}

// defaultFixtures are the grok and codex usage payloads: a monthly grok
// allowance 73% used and 62% of the codex limit used.
func defaultFixtures() map[string]string {
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

// commandCodeBodies is the Command Code account the fake API describes: mid
// period with 3 days to renewal, so the projected renewal counter is stable.
func commandCodeBodies() map[string]string {
now := time.Now()
periodStart := now.Add(-7 * 24 * time.Hour).Format(time.RFC3339)
periodEnd := now.Add(3 * 24 * time.Hour).Format(time.RFC3339)
weeklyReset := now.Add(2 * 24 * time.Hour).UnixMilli()
return map[string]string{
commandcode.PathWhoami: `{"success":true,"user":{"id":"u1","userName":"alice","name":"Alice","email":"alice@example.test"}}`,
commandcode.PathCredits: `{"credits":{"belowThreshold":false,"creditThreshold":0,"monthlyCredits":7.5,` +
`"purchasedCredits":0,"freeCredits":0},"windowLimits":{"limited":true,` +
`"fiveHour":{"used":1.5,"cap":3,"exceeded":false,"resetAt":0},` +
`"weekly":{"used":1.5,"cap":6,"exceeded":false,"resetAt":` + fmt.Sprintf("%d", weeklyReset) + `}}}`,
commandcode.PathSubscriptions: `{"success":true,"data":{"id":"sub_1","status":"active","planId":"individual-go",` +
`"quantity":1,"currentPeriodStart":"` + periodStart + `","currentPeriodEnd":"` + periodEnd + `"}}`,
commandcode.PathUsageSummary: `{"totalCount":120,"totalCost":2.5,"averageCost":0.020833333,` +
`"successRate":96,"completedCount":115,"failedCount":5,"totalTokens":1000,"totalTokensIn":600,` +
`"totalTokensOut":400,"totalCredits":2.5,"totalMonthlyCredits":2.5,"periodBasis":"billing-period"}`,
}
}

// SeedSessions gives every provider home one session, so a snapshot's session
// totals are predictable from any leaf.
func SeedSessions(t *testing.T, req *Request) {
t.Helper()
dir := filepath.Join(req.GrokHome, "sessions", "2026-09-10T09-00-00-main")
if err := os.MkdirAll(dir, 0o755); err != nil {
t.Fatalf("mkdir grok session: %v", err)
}
summary := `{"created_at":"2026-09-10T09:00:00Z","updated_at":"2026-09-10T10:00:00Z"}`
if err := os.WriteFile(filepath.Join(dir, "summary.json"), []byte(summary), 0o644); err != nil {
t.Fatalf("write grok summary: %v", err)
}

codexDir := filepath.Join(req.CodexHome, "sessions", "2026", "09", "10")
if err := os.MkdirAll(codexDir, 0o755); err != nil {
t.Fatalf("mkdir codex session: %v", err)
}
rollout := filepath.Join(codexDir, "rollout-2026-09-10T09-00-00-11111111-1111-4111-8111-111111111111.jsonl")
if err := os.WriteFile(rollout, []byte(`{"type":"session_meta"}`+"\n"), 0o644); err != nil {
t.Fatalf("write codex rollout: %v", err)
}

ccDir := filepath.Join(req.CommandCodeHome, "projects", "-Users-tester-project")
if err := os.MkdirAll(ccDir, 0o755); err != nil {
t.Fatalf("mkdir commandcode project: %v", err)
}
transcript := filepath.Join(ccDir, "22222222-2222-4222-8222-222222222222.jsonl")
if err := os.WriteFile(transcript, []byte(`{"type":"session","timestamp":"2026-09-10T09:00:00Z"}`+"\n"), 0o644); err != nil {
t.Fatalf("write commandcode transcript: %v", err)
}
}

// WriteSnapshot plants one stored snapshot line, for leaves that need history
// without running a collect first.
func WriteSnapshot(t *testing.T, home string, provider usage.ProviderID, day, name string, ts time.Time, usedPercent, sessions int) string {
t.Helper()
dir := filepath.Join(home, "usages", string(provider), day)
if err := os.MkdirAll(dir, 0o755); err != nil {
t.Fatalf("mkdir snapshot dir: %v", err)
}
record := usage.Record{
Schema:   usage.SchemaID,
TS:       ts.UTC(),
Provider: provider,
Usage: usage.UsageBlock{
OK:       true,
Endpoint: "billing",
Values:   map[string]float64{"used_percent": float64(usedPercent)},
Display:  map[string]string{"usage": fmt.Sprintf("%d%% used", usedPercent)},
},
Sessions: usage.SessionsBlock{Total: sessions},
}
line, err := record.MarshalJSONLine()
if err != nil {
t.Fatalf("encode snapshot: %v", err)
}
path := filepath.Join(dir, name+usage.SnapshotSuffix)
if err := os.WriteFile(path, append(line, '\n'), 0o644); err != nil {
t.Fatalf("write snapshot: %v", err)
}
return path
}

// AssertExitCode fails unless the CLI exited with want.
func AssertExitCode(t *testing.T, resp *Response, want int) {
t.Helper()
if resp.ExitCode != want {
t.Fatalf("exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s",
resp.ExitCode, want, resp.Stdout, resp.Stderr)
}
}

// AssertStderrContains fails unless stderr contains want.
func AssertStderrContains(t *testing.T, resp *Response, want string) {
t.Helper()
if !strings.Contains(resp.Stderr, want) {
t.Fatalf("stderr does not contain %q:\n%s", want, resp.Stderr)
}
}

// AssertStdoutContains fails unless stdout contains want.
func AssertStdoutContains(t *testing.T, resp *Response, want string) {
t.Helper()
if !strings.Contains(resp.Stdout, want) {
t.Fatalf("stdout does not contain %q:\n%s", want, resp.Stdout)
}
}

// AssertNoSnapshots fails when any snapshot file exists.
func AssertNoSnapshots(t *testing.T, resp *Response) {
t.Helper()
if len(resp.Files) != 0 {
t.Fatalf("unexpected snapshot files: %v", resp.Files)
}
}

// AssertSessionsInStore checks one stored record's session total and that the
// snapshot's day directory is today in local time.
func AssertStoreDay(t *testing.T, home string, provider usage.ProviderID, path string) {
t.Helper()
wantDay := time.Now().Local().Format("2006-01-02")
if got := filepath.Base(filepath.Dir(path)); got != wantDay {
t.Fatalf("%s snapshot day dir = %s, want %s", provider, got, wantDay)
}
if got := filepath.Base(filepath.Dir(filepath.Dir(path))); got != string(provider) {
t.Fatalf("snapshot provider dir = %s, want %s", got, provider)
}
}

// runViewHTTP probes a `usage view` server a view leaf started in Setup. The
// server's output so far comes back with the response, so an assertion can check
// the HTTP answer and what the CLI printed at startup together.
func runViewHTTP(t *testing.T, req *Request) (*Response, error) {
t.Helper()
path := req.HTTPPath
if path == "" {
path = "/api/healthz"
}
stdout, stderr := viewServerOutput(req)
client := &http.Client{Timeout: 10 * time.Second}
httpResp, err := client.Get(req.ViewBaseURL + path)
if err != nil {
return &Response{Stdout: stdout, Stderr: stderr}, fmt.Errorf("GET %s: %w", path, err)
}
defer httpResp.Body.Close()
body, err := io.ReadAll(httpResp.Body)
if err != nil {
return nil, fmt.Errorf("read %s: %w", path, err)
}
return &Response{
Stdout:     stdout,
Stderr:     stderr,
HTTPStatus: httpResp.StatusCode,
HTTPBody:   string(body),
}, nil
}

// viewServerOutput is the view server's stdout and stderr so far.
func viewServerOutput(req *Request) (string, string) {
var stdout, stderr string
if req.ViewStdout != nil {
stdout = req.ViewStdout()
}
if req.ViewStderr != nil {
stderr = req.ViewStderr()
}
return stdout, stderr
}

// AssertHTTPStatus fails unless the probe answered with want.
func AssertHTTPStatus(t *testing.T, resp *Response, want int) {
t.Helper()
if resp.HTTPStatus != want {
t.Fatalf("HTTP status = %d, want %d\nbody:\n%s", resp.HTTPStatus, want, resp.HTTPBody)
}
}

// DecodeJSON decodes a whole JSON body into target.
func DecodeJSON(t *testing.T, body string, target any) {
t.Helper()
if err := json.Unmarshal([]byte(body), target); err != nil {
t.Fatalf("decode JSON body: %v\n%s", err, body)
}
}

// AssertStoreUnchanged fails unless the store on disk still holds exactly the
// files the leaf planted, with exactly the lines it wrote: viewing never writes.
func AssertStoreUnchanged(t *testing.T, req *Request) {
t.Helper()
found := map[string]string{}
err := filepath.WalkDir(filepath.Join(req.Home, "usages"), func(path string, entry fs.DirEntry, err error) error {
if err != nil {
return err
}
if entry.IsDir() {
return nil
}
content, err := os.ReadFile(path)
if err != nil {
return err
}
found[path] = string(content)
return nil
})
if err != nil && !os.IsNotExist(err) {
t.Fatalf("read the store: %v", err)
}
if len(found) != len(req.ViewFiles) {
t.Fatalf("store holds %d files, want the %d that were planted: %v", len(found), len(req.ViewFiles), found)
}
for path, want := range req.ViewFiles {
got, ok := found[path]
if !ok {
t.Fatalf("planted snapshot %s is gone", path)
}
if got != want {
t.Fatalf("snapshot %s changed:\n got: %q\nwant: %q", path, got, want)
}
}
}
```
