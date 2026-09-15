# agent-pro usage CLI tests

Doc-style tests for the `agent-pro usage` command line: `collect` writes one
snapshot per provider under `$AGENT_PRO_HOME/usages`, `list` reads them back,
`--cron` repeats the cycle on a schedule, and `view` serves the stored snapshots
as a read-only dashboard.

# DSN (Domain Specific Notion)

```
agent-pro usage collect [--json] [--cron SPEC] [--max-runs N] [--timeout DUR]
                        [--grok-home D] [--codex-home D] [--commandcode-home D]
agent-pro usage list [--last N] [--json]
agent-pro usage view [--port N] [--open|--no-open] [--since DUR]
                     [--static-dir D] [--json]
```

**Participants**

- **agent-pro binary** — built from `cmd/agent-pro` into the leaf's temp dir, so
  these leaves exercise the real flag parsing, exit status and stream split.
- **Store** — `$AGENT_PRO_HOME/usages/<provider>/<local date>/<local time>-snapshot.jsonl`,
  one JSON record per provider per cycle.
- **Provider homes** — credentials for the usage fetch and sessions for the
  session block, selected by `--grok-home` / `--codex-home` /
  `--commandcode-home`.
- **Fixture transport** — `AGENT_PRO_USAGE_FIXTURE_DIR` serves grok and codex from
  JSON files; `COMMANDCODE_SANDBOX` + `COMMANDCODE_API_URL` point Command Code at
  the fake API the harness starts.
- **Cron loop** — `pkgs/cronspec`: a 5-field spec, `@hourly|@daily|@midnight|@weekly|@monthly`
  or `@every <duration>`; `--max-runs` bounds a run, and each wait logs
  `usage collect: next run at <next fire time>` on stderr, in local time like
  the schedule itself.
- **Dashboard server** — `agent-pro usage view` binds `127.0.0.1`, starting at 8080
  and skipping busy ports (`--port 0` = any free port). It prints `serving <url>`
  first on stdout, then the store line; `/api/summary`, `/api/series` and
  `/api/healthz` answer over HTTP. View leaves start it in the background, probe one
  path, and kill it, so they assert HTTP plus the startup output instead of an exit
  status.

```
collect -> fetch three providers -> append three snapshot files
           stderr: "warning: <provider>: <error>" per failed fetch
           exit 0 if at least one provider answered, else exit 1
cron    -> cycle immediately, then per schedule until --max-runs or SIGINT
list    -> newest-first table, or raw records with --json
view    -> read the store once, print the URL, serve until killed
           stdout: "serving http://127.0.0.1:<port>" then the store summary
           stderr: "warning: <reason>" per skipped file, then the stop hint
           HTTP:   / -> page, /api/summary, /api/series, /api/healthz
doctest <- stdout, stderr, exit status, HTTP bodies, and the files under $AGENT_PRO_HOME/usages
```

## Version

0.0.1

## Decision Tree

```
cmd/agent-pro/tests/usage/
├── DOCTEST.md
├── SETUP.md
├── collect/                     # one collect cycle per leaf
│   ├── json-three-providers/    # --json: three records, three files, exit 0
│   ├── failure-warns-exit-0/    # grok + codex fail: warnings, records kept, exit 0
│   └── all-failed-exits-1/      # no provider answers: exit 1, records still written
├── cron/
│   ├── max-runs-appends/        # @every + --max-runs: one cycle per run
│   └── invalid-spec-exits-1/    # a spec that is not a schedule: exit 1
├── list/
│   └── newest-first/            # stored records newest first, --last truncation
└── view/                        # the dashboard server, one HTTP probe per leaf
    ├── index-serves/            # GET / -> the embedded page, URL line on stdout
    ├── summary-three-providers/ # newest snapshot per provider, recent, metrics
    ├── series-window/           # --since window, ascending points, step counters
    ├── series-unknown-metric/   # unknown metric -> 200 with an empty series
    ├── empty-store/             # no snapshots: 200, empty arrays, exit 0
    ├── port-busy-falls-back/    # busy --port: warning on stderr, next free port
    ├── read-only/               # viewing never writes to the store
    ├── corrupt-snapshot-warns/  # unparseable file: warning, other providers chart
    ├── invalid-port-exits-1/    # --port -1: exit 1, nothing served
    ├── help-lists-view/         # usage --help lists view
    └── view-help/               # usage view --help documents the flags
```

Parameter ranking (most → least significant):

1. **Exit status** — 0 while any provider answers, 1 when none does or the spec is invalid
2. **Snapshot files** — one per provider per cycle, appended under the provider/day tree
3. **Stream split** — JSON or table on stdout, warnings and cron logs on stderr
4. **Cron loop** — cycle count from `--max-runs`
5. **Read back** — `usage list` ordering and `--last`
6. **Dashboard HTTP** — `/api/summary` and `/api/series` payloads, and that serving writes nothing

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `collect/json-three-providers` | `--json` prints three records, one file per provider, exit 0 |
| 2 | `collect/failure-warns-exit-0` | Two failed fetches warn on stderr, keep their records, exit 0 |
| 3 | `collect/all-failed-exits-1` | Every fetch fails: exit 1 with `no usage collected` |
| 4 | `cron/max-runs-appends` | `--cron '@every ...' --max-runs 3` appends three cycles |
| 5 | `cron/invalid-spec-exits-1` | `--cron 'every 5 min'` reports the bad spec and exits 1 |
| 6 | `list/newest-first` | `usage list` shows the newest snapshots first, `--last` truncates |
| 7 | `view/index-serves` | `GET /` returns the dashboard page; stdout carries the bound URL |
| 8 | `view/summary-three-providers` | One entry per provider with its newest snapshot, plus recent and metrics |
| 9 | `view/series-window` | `since=24h` drops the older point; counters are marked as steps |
| 10 | `view/series-unknown-metric` | An unknown metric answers 200 with an empty series |
| 11 | `view/empty-store` | No snapshots: 200 with empty arrays and a `0 snapshots` store line |
| 12 | `view/port-busy-falls-back` | An explicit busy port warns and serves on the next free one |
| 13 | `view/read-only` | Serving leaves every snapshot file byte-identical |
| 14 | `view/corrupt-snapshot-warns` | A broken line warns, is skipped, and is reported in `errors` |
| 15 | `view/invalid-port-exits-1` | `--port -1` exits 1 with the range message and empty stdout |
| 16 | `view/help-lists-view` | `usage --help` lists `view` |
| 17 | `view/view-help` | `usage view --help` documents the flags, the metrics and the exit contract |

## How to Run

```sh
doctest vet ./cmd/agent-pro/tests/usage
doctest test -v ./cmd/agent-pro/tests/usage
doctest test -v ./cmd/agent-pro/tests/usage/collect/json-three-providers
```

```go
import (
"bytes"
"context"
"encoding/json"
"errors"
"io/fs"
"net/http"
"net/http/httptest"
"os"
"os/exec"
"path/filepath"
"sort"
"strings"
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
"github.com/xhd2015/doctest/session"
)

type Request struct {
// Bin is the agent-pro binary built by the root Setup.
Bin string
// Command is the leading arguments; empty means "usage collect".
Command []string
// Args are the remaining arguments.
Args []string
// WithHomes adds the three --*-home flags (collect only).
WithHomes bool
// Home is AGENT_PRO_HOME, the parent of the store root.
Home string
// GrokHome, CodexHome and CommandCodeHome hold credentials and sessions.
GrokHome        string
CodexHome       string
CommandCodeHome string
// BareCommandCodeHome has no credentials, so a Command Code fetch fails locally.
BareCommandCodeHome string
// FixtureDir serves grok and codex usage; an empty fixture directory makes both
// fail without touching the network.
FixtureDir string
// EmptyFixtureDir is a directory that holds no fixtures.
EmptyFixtureDir string
// ServeCommandCode starts the fake Command Code API and points the CLI at it.
ServeCommandCode bool
// Timeout bounds the CLI run.
Timeout time.Duration
// Mode selects the harness path: "" runs the CLI in the foreground, "view"
// probes a `usage view` server that the view group's Setup started.
Mode string
// HTTPPath is the probe path in Mode "view"; empty means /api/healthz.
HTTPPath string
// ViewPort is the port a view leaf passes to --port; 0 asks for any free one.
ViewPort int
// ViewArgs are extra arguments after `usage view`.
ViewArgs []string
// ViewBaseURL is the running view server's URL; ViewStdout and ViewStderr read
// the output that server has printed so far, and ViewCmd is the process itself.
ViewBaseURL string
ViewStdout  func() string
ViewStderr  func() string
ViewCmd     *exec.Cmd
// ViewFiles maps a snapshot file a view leaf planted to the exact line it
// wrote, so an assertion can prove that viewing changed nothing on disk.
ViewFiles map[string]string
}

type Response struct {
Stdout   string
Stderr   string
ExitCode int
// Files are every snapshot file under the store, sorted by path.
Files []string
// Lines holds each snapshot file's records as raw lines.
Lines map[string][]string
// HTTPStatus and HTTPBody are the probe result in Mode "view".
HTTPStatus int
HTTPBody   string
}

// Records returns every stored record, in file order.
func (r *Response) Records(t *testing.T) []usage.Record {
t.Helper()
var out []usage.Record
for _, path := range r.Files {
for _, line := range r.Lines[path] {
out = append(out, DecodeRecord(t, line))
}
}
return out
}

// RecordsFor returns the stored records of one provider.
func (r *Response) RecordsFor(t *testing.T, provider usage.ProviderID) []usage.Record {
t.Helper()
var out []usage.Record
for _, path := range r.Files {
if filepath.Base(filepath.Dir(filepath.Dir(path))) != string(provider) {
continue
}
for _, line := range r.Lines[path] {
out = append(out, DecodeRecord(t, line))
}
}
return out
}

// StdoutLines returns the non-empty stdout lines.
func (r *Response) StdoutLines() []string {
return NonEmptyLines(r.Stdout)
}

// StderrLines returns the non-empty stderr lines.
func (r *Response) StderrLines() []string {
return NonEmptyLines(r.Stderr)
}

// Run dispatches on the mode: the default path executes the CLI once, and the
// view path probes the server the view group's Setup already started.
func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
t.Helper()
if req.Mode == "view" {
return runViewHTTP(t, req)
}
return runUsageCLI(t, req)
}

func runUsageCLI(t *testing.T, req *Request) (*Response, error) {
t.Helper()
command := req.Command
if len(command) == 0 {
command = []string{"usage", "collect"}
}
args := append([]string{}, command...)
if req.WithHomes {
args = append(args,
"--grok-home", req.GrokHome,
"--codex-home", req.CodexHome,
"--commandcode-home", req.CommandCodeHome)
}
args = append(args, req.Args...)

env := CLIEnv(req)
if req.ServeCommandCode {
server := ServeCommandCode(t)
env = append(env, "COMMANDCODE_SANDBOX=true", "COMMANDCODE_API_URL="+server.URL)
}

timeout := req.Timeout
if timeout <= 0 {
timeout = 60 * time.Second
}
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()

cmd := exec.CommandContext(ctx, req.Bin, args...)
cmd.Env = env
var stdout, stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr
runErr := cmd.Run()

resp := &Response{
Stdout: stdout.String(),
Stderr: stderr.String(),
Lines:  map[string][]string{},
}
if ctx.Err() != nil {
return resp, ctx.Err()
}
if runErr != nil {
var exitErr *exec.ExitError
if !errors.As(runErr, &exitErr) {
return resp, runErr
}
resp.ExitCode = exitErr.ExitCode()
}

resp.Files, resp.Lines = ReadSnapshots(t, filepath.Join(req.Home, "usages"))
return resp, nil
}

// CLIEnv is the child environment: the harness's values win, and the parent's
// provider credentials can never leak into a leaf.
func CLIEnv(req *Request) []string {
blocked := map[string]bool{
"AGENT_PRO_HOME":              true,
"GROK_HOME":                   true,
"CODEX_HOME":                  true,
"AGENT_PRO_USAGE_FIXTURE_DIR": true,
"COMMANDCODE_SANDBOX":         true,
"COMMANDCODE_API_URL":         true,
"COMMAND_CODE_API_KEY":        true,
}
env := make([]string, 0, len(os.Environ())+4)
for _, entry := range os.Environ() {
key, _, _ := strings.Cut(entry, "=")
if blocked[key] {
continue
}
env = append(env, entry)
}
return append(env,
"AGENT_PRO_HOME="+req.Home,
"GROK_HOME="+req.GrokHome,
"CODEX_HOME="+req.CodexHome,
"AGENT_PRO_USAGE_FIXTURE_DIR="+req.FixtureDir,
)
}

// ServeCommandCode starts the fake Command Code API the payloads describe.
func ServeCommandCode(t *testing.T) *httptest.Server {
t.Helper()
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
body, ok := commandCodeBodies()[r.URL.Path]
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

// ReadSnapshots lists every snapshot file under root with its raw lines. A root
// that does not exist yields nothing.
func ReadSnapshots(t *testing.T, root string) ([]string, map[string][]string) {
t.Helper()
files := []string{}
lines := map[string][]string{}
err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
if err != nil {
return err
}
if entry.IsDir() || !strings.HasSuffix(entry.Name(), usage.SnapshotSuffix) {
return nil
}
content, err := os.ReadFile(path)
if err != nil {
return err
}
files = append(files, path)
lines[path] = NonEmptyLines(string(content))
return nil
})
if err != nil {
if !os.IsNotExist(err) {
t.Fatalf("read snapshots under %s: %v", root, err)
}
return nil, lines
}
sort.Strings(files)
return files, lines
}

// DecodeRecord decodes one snapshot line.
func DecodeRecord(t *testing.T, line string) usage.Record {
t.Helper()
var record usage.Record
if err := json.Unmarshal([]byte(line), &record); err != nil {
t.Fatalf("decode snapshot record: %v\n%s", err, line)
}
return record
}

// DecodeJSONLines decodes every line of a --json output.
func DecodeJSONLines(t *testing.T, output string) []usage.Record {
t.Helper()
lines := NonEmptyLines(output)
records := make([]usage.Record, 0, len(lines))
for _, line := range lines {
records = append(records, DecodeRecord(t, line))
}
return records
}

// NonEmptyLines splits text into its non-empty lines.
func NonEmptyLines(text string) []string {
var out []string
for _, line := range strings.Split(text, "\n") {
if strings.TrimSpace(line) != "" {
out = append(out, line)
}
}
return out
}

// ProviderDir is the store directory holding one provider's snapshots.
func ProviderDir(home string, provider usage.ProviderID) string {
return filepath.Join(home, "usages", string(provider))
}
```
