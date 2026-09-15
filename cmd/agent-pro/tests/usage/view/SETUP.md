# Scenario

**Feature**: `usage view` serves the stored snapshots as a read-only dashboard

```
usage view [--port N] [--open|--no-open] [--since DUR] [--json]
stdout: serving http://127.0.0.1:<port>
        store   <root>  (read-only, 3 providers, 4 snapshots)
GET /                -> the dashboard page
GET /api/summary     -> latest snapshot per provider, recent records, metric menu
GET /api/series?...  -> per-provider series for the requested metrics
GET /api/healthz     -> ok
```

## Preconditions

- The store is a temp `$AGENT_PRO_HOME/usages`; leaves plant snapshots with
  `WriteViewSnapshot` instead of running a collect, so the charted values are fixed.
- A server leaf starts the real binary in the background and reads the URL line from
  its stdout; the harness probes it over HTTP and kills it at cleanup.
- Ports: leaves ask for `--port 0` (any free port) except the fallback leaf, which
  holds a specific port itself. No leaf needs port 8080, so the suite never
  contends for it.

## Steps

1. Group `Setup` selects the view mode (server leaves never run the CLI in the
   foreground); the two CLI leaves reset `Mode` to `""`.
2. Leaf `Setup` plants snapshots, picks `HTTPPath`, and calls `startViewServer`.
3. `Run` performs the HTTP probe and returns the server's output with the response.
4. `Assert` checks the HTTP payload, the startup output, and the untouched store.

```go
import (
"bufio"
"bytes"
"io"
"net"
"net/http"
"os"
"os/exec"
"path/filepath"
"strconv"
"strings"
"sync"
"testing"
"time"

"github.com/xhd2015/agent-pro/agent/usage"
"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
req.Mode = "view"
return nil
}

// viewServer is a background `usage view` process with its output so far.
// Output() and Errors() read it; the fields stay unexported here because the
// harness exports them when it assembles this file.
type viewServer struct {
mu     sync.Mutex
stdout bytes.Buffer
stderr bytes.Buffer
}

func (s *viewServer) Output() string {
s.mu.Lock()
defer s.mu.Unlock()
return s.stdout.String()
}

func (s *viewServer) Errors() string {
s.mu.Lock()
defer s.mu.Unlock()
return s.stderr.String()
}

// startViewServer launches `usage view` for the leaf and waits until it serves:
// the URL line on stdout, then a 200 from /api/healthz. The process is killed at
// cleanup, so its exit status is never asserted — a server is stopped, not ended.
func startViewServer(t *testing.T, req *Request) {
t.Helper()
server := &viewServer{}
args := []string{"usage", "view", "--no-open", "--port", strconv.Itoa(req.ViewPort)}
args = append(args, req.ViewArgs...)

cmd := exec.Command(req.Bin, args...)
cmd.Env = CLIEnv(req)
stdoutPipe, err := cmd.StdoutPipe()
if err != nil {
t.Fatalf("stdout pipe: %v", err)
}
stderrPipe, err := cmd.StderrPipe()
if err != nil {
t.Fatalf("stderr pipe: %v", err)
}
if err := cmd.Start(); err != nil {
t.Fatalf("start usage view: %v", err)
}
req.ViewCmd = cmd
t.Cleanup(func() {
if cmd.Process != nil {
_ = cmd.Process.Kill()
}
_ = cmd.Wait()
})

go drainPipe(stdoutPipe, &server.mu, &server.stdout)
go drainPipe(stderrPipe, &server.mu, &server.stderr)
req.ViewStdout = server.Output
req.ViewStderr = server.Errors

req.ViewBaseURL = waitForServingURL(t, server)
if !waitForHealth(t, req.ViewBaseURL+"/api/healthz") {
t.Fatalf("view server never answered /api/healthz\nstdout:\n%s\nstderr:\n%s",
server.Output(), server.Errors())
}
}

func drainPipe(reader io.Reader, mu *sync.Mutex, target *bytes.Buffer) {
scanner := bufio.NewScanner(reader)
scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
for scanner.Scan() {
mu.Lock()
target.WriteString(scanner.Text())
target.WriteString("\n")
mu.Unlock()
}
}

func waitForServingURL(t *testing.T, server *viewServer) string {
t.Helper()
deadline := time.Now().Add(10 * time.Second)
for time.Now().Before(deadline) {
for _, line := range strings.Split(server.Output(), "\n") {
if strings.HasPrefix(line, "serving http://") {
return strings.TrimSpace(strings.TrimPrefix(line, "serving "))
}
}
time.Sleep(20 * time.Millisecond)
}
t.Fatalf("no serving line on stdout\nstdout:\n%s\nstderr:\n%s", server.Output(), server.Errors())
return ""
}

func waitForHealth(t *testing.T, url string) bool {
t.Helper()
client := &http.Client{Timeout: 5 * time.Second}
deadline := time.Now().Add(10 * time.Second)
for time.Now().Before(deadline) {
resp, err := client.Get(url)
if err == nil {
_ = resp.Body.Close()
if resp.StatusCode == http.StatusOK {
return true
}
}
time.Sleep(20 * time.Millisecond)
}
return false
}

// WriteViewSnapshot plants one stored snapshot line whose values the leaf picks,
// so the dashboard sees exactly the metrics it is meant to chart. It records the
// exact bytes in req.ViewFiles for the read-only assertion and returns the file
// path with the line written.
func WriteViewSnapshot(t *testing.T, req *Request, provider usage.ProviderID, ts time.Time,
values map[string]float64, display map[string]string, sessions int, ok bool, usageErr string) (string, string) {
t.Helper()
record := usage.Record{
Schema:   usage.SchemaID,
TS:       ts.UTC(),
Provider: provider,
Usage: usage.UsageBlock{
OK:       ok,
Endpoint: "billing",
Values:   values,
Display:  display,
Error:    usageErr,
},
Sessions: usage.SessionsBlock{Total: sessions},
}
line, err := record.MarshalJSONLine()
if err != nil {
t.Fatalf("encode %s snapshot: %v", provider, err)
}
dir := filepath.Join(req.Home, "usages", string(provider), ts.Local().Format("2006-01-02"))
if err := os.MkdirAll(dir, 0o755); err != nil {
t.Fatalf("mkdir snapshot dir: %v", err)
}
path := filepath.Join(dir, ts.Local().Format("15-04-05")+usage.SnapshotSuffix)
body := append(line, '\n')
if err := os.WriteFile(path, body, 0o644); err != nil {
t.Fatalf("write %s snapshot: %v", provider, err)
}
if req.ViewFiles == nil {
req.ViewFiles = map[string]string{}
}
req.ViewFiles[path] = string(body)
return path, string(line)
}

// Display is the display block the planted snapshots carry.
func Display(headline, detail string) map[string]string {
return map[string]string{"usage": headline, "detail": detail}
}

// AssertServingURL checks the first stdout line is the URL the harness probed.
func AssertServingURL(t *testing.T, resp *Response, req *Request) {
t.Helper()
lines := resp.StdoutLines()
if len(lines) == 0 {
t.Fatalf("stdout is empty, want the serving line first")
}
const want = "serving http://127.0.0.1:"
if !strings.HasPrefix(lines[0], want) {
t.Fatalf("stdout line 1 = %q, want it to start with %q", lines[0], want)
}
if got := strings.TrimPrefix(lines[0], "serving "); got != req.ViewBaseURL {
t.Fatalf("serving URL = %q, want the probed %q", got, req.ViewBaseURL)
}
}

// AssertStoreSummaryLine checks the store line names the root and the counts.
func AssertStoreSummaryLine(t *testing.T, resp *Response, want string) {
t.Helper()
for _, line := range resp.StdoutLines() {
if strings.HasPrefix(line, "store ") {
if !strings.Contains(line, want) {
t.Fatalf("store line = %q, want it to contain %q", line, want)
}
return
}
}
t.Fatalf("no store line on stdout:\n%s", resp.Stdout)
}

// HoldPort occupies a local port for the fallback leaf and returns it with its
// release function.
func HoldPort(t *testing.T) (int, func()) {
t.Helper()
listener, err := net.Listen("tcp", "127.0.0.1:0")
if err != nil {
t.Fatalf("hold a port: %v", err)
}
port := listener.Addr().(*net.TCPAddr).Port
return port, func() { _ = listener.Close() }
}

// ProviderJSON is one entry of the /api/summary providers array.
type ProviderJSON struct {
Provider   string  `json:"provider"`
AgeSeconds float64 `json:"age_seconds"`
Count      int     `json:"count"`
Latest     struct {
TS       time.Time `json:"ts"`
Provider string    `json:"provider"`
Usage    struct {
OK      bool               `json:"ok"`
Values  map[string]float64 `json:"values"`
Display map[string]string  `json:"display"`
Error   string             `json:"error"`
} `json:"usage"`
Sessions struct {
Total int `json:"total"`
} `json:"sessions"`
Path string `json:"path"`
} `json:"latest"`
}

// MetricJSON is one entry of the /api/summary metrics array.
type MetricJSON struct {
Name      string   `json:"name"`
Kind      string   `json:"kind"`
Unit      string   `json:"unit"`
Providers []string `json:"providers"`
Count     int      `json:"count"`
Chartable bool     `json:"chartable"`
}

// RecordJSON is one entry of the /api/summary recent array, with its file.
type RecordJSON struct {
TS       time.Time `json:"ts"`
Provider string    `json:"provider"`
Usage    struct {
OK      bool               `json:"ok"`
Values  map[string]float64 `json:"values"`
Display map[string]string  `json:"display"`
Error   string             `json:"error"`
} `json:"usage"`
Sessions struct {
Total int `json:"total"`
} `json:"sessions"`
Path string `json:"path"`
}

// SummaryJSON is the /api/summary payload.
type SummaryJSON struct {
Schema        string         `json:"schema"`
GeneratedAt   time.Time      `json:"generated_at"`
Root          string         `json:"root"`
Range         string         `json:"range"`
SnapshotCount int            `json:"snapshot_count"`
Providers     []ProviderJSON `json:"providers"`
Recent        []RecordJSON   `json:"recent"`
Metrics       []MetricJSON   `json:"metrics"`
Errors        []struct {
Path  string `json:"path"`
Error string `json:"error"`
} `json:"errors"`
}

// PointJSON is one charted value.
type PointJSON struct {
TS    time.Time `json:"ts"`
Value float64   `json:"value"`
}

// SeriesJSON is one provider's series.
type SeriesJSON struct {
Provider string      `json:"provider"`
Points   []PointJSON `json:"points"`
Latest   float64     `json:"latest"`
Min      float64     `json:"min"`
Max      float64     `json:"max"`
}

// MetricBlockJSON is one metric in the /api/series payload.
type MetricBlockJSON struct {
Name   string       `json:"name"`
Unit   string       `json:"unit"`
Kind   string       `json:"kind"`
Step   bool         `json:"step"`
Series []SeriesJSON `json:"series"`
}

// SeriesJSONPayload is the /api/series payload.
type SeriesJSONPayload struct {
Schema  string            `json:"schema"`
Range   string            `json:"range"`
Metrics []MetricBlockJSON `json:"metrics"`
}

// Summary decodes the probed /api/summary body.
func Summary(t *testing.T, resp *Response) SummaryJSON {
t.Helper()
var summary SummaryJSON
DecodeJSON(t, resp.HTTPBody, &summary)
return summary
}

// Series decodes the probed /api/series body.
func Series(t *testing.T, resp *Response) SeriesJSONPayload {
t.Helper()
var payload SeriesJSONPayload
DecodeJSON(t, resp.HTTPBody, &payload)
return payload
}

// Provider returns one provider's summary entry.
func Provider(t *testing.T, summary SummaryJSON, name string) ProviderJSON {
t.Helper()
for _, entry := range summary.Providers {
if entry.Provider == name {
return entry
}
}
t.Fatalf("%s is missing from the summary providers %v", name, summary.Providers)
return ProviderJSON{}
}

// MetricBlock returns one metric block of a /api/series payload.
func MetricBlock(t *testing.T, payload SeriesJSONPayload, name string) MetricBlockJSON {
t.Helper()
for _, block := range payload.Metrics {
if block.Name == name {
return block
}
}
t.Fatalf("metric %q is missing from %v", name, payload.Metrics)
return MetricBlockJSON{}
}

// SeriesFor returns one provider's series inside a metric block.
func SeriesFor(t *testing.T, block MetricBlockJSON, provider string) SeriesJSON {
t.Helper()
for _, series := range block.Series {
if series.Provider == provider {
return series
}
}
t.Fatalf("%s is missing from metric %q series %v", provider, block.Name, block.Series)
return SeriesJSON{}
}

// AssertNoWarnings fails when the server warned, so a clean leaf stays clean.
func AssertNoWarnings(t *testing.T, resp *Response) {
t.Helper()
if strings.Contains(resp.Stderr, "warning:") {
t.Fatalf("stderr = %q, want no warnings", resp.Stderr)
}
}
```
