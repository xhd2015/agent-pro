package view

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/xhd2015/agent-pro/agent/usage"
)

// DefaultPort is the first port `usage view` tries. A busy port is skipped for
// up to viewPortAttempts ports.
const (
	DefaultPort      = 8080
	viewPortAttempts = 100
)

// DefaultSince is the chart range used when --since is not given.
const DefaultSince = 7 * 24 * time.Hour

// SchemaID is the schema of the dashboard's JSON API.
const SchemaID = "agent-pro/usage-view/v1"

// RecentLimit is how many newest snapshots /api/summary carries for its table.
const RecentLimit = 50

//go:embed dashboard.html
var dashboardHTML []byte

// Options configures the read-only dashboard server.
type Options struct {
	// Port is the preferred listen port; zero or negative binds any free port.
	Port int
	// PortExplicit records that --port was given, so falling back to the next
	// free port is worth a warning.
	PortExplicit bool
	// Since is the default chart range.
	Since time.Duration
	// StaticDir serves dashboard.html from disk instead of the embedded copy.
	StaticDir string
	// Open opens the dashboard in a browser once it is listening.
	Open bool
	// JSON prints one machine-readable line for the bound URL instead of the
	// human summary.
	JSON bool
	// Stdout receives the result lines. Defaults to os.Stdout.
	Stdout io.Writer
	// Stderr receives warnings and the stop hint. Defaults to os.Stderr.
	Stderr io.Writer
	// Now overrides the clock, for tests.
	Now func() time.Time
}

func (o Options) now() time.Time {
	if o.Now != nil {
		return o.Now()
	}
	return time.Now()
}

// Serve starts the dashboard over store and blocks until ctx is cancelled, a
// signal arrives, or the listener fails. It only reads the store.
func Serve(ctx context.Context, store *usage.Store, opts Options) error {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	if opts.Since <= 0 {
		opts.Since = DefaultSince
	}

	// The store is read before binding, so its warnings and summary are on the
	// streams before the URL line a caller waits for.
	records, failures, err := store.ListFiltered(usage.ListOptions{})
	if err != nil {
		return fmt.Errorf("read %s: %w", store.Root, err)
	}
	for _, failure := range failures {
		fmt.Fprintf(opts.Stderr, "warning: usage view: skipping %s: %v\n",
			displayPath(failure.Path, homeDir()), failure.Err)
	}

	listener, err := listenLocalPreferred(opts.Port)
	if err != nil {
		return err
	}
	defer listener.Close()
	port := listenerPort(listener)

	printStartup(opts, store, records, len(failures), port)
	if opts.Open {
		openLocalURL(urlFor(port))
	}
	fmt.Fprintln(opts.Stderr, "usage view: press Ctrl+C to stop")

	server := &http.Server{Handler: Handler(store, opts)}
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()

	signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	case <-signalCtx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
		return nil
	}
}

// Handler serves the dashboard page and its read-only JSON API over one store.
// The store is re-read per request, so a new collect shows up on the next poll.
func Handler(store *usage.Store, opts Options) http.Handler {
	opts = normalizeOptions(opts)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		page, err := dashboardPage(opts.StaticDir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(page)
	})
	mux.HandleFunc("/api/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeText(w, http.StatusOK, "ok")
	})
	mux.HandleFunc("/api/summary", func(w http.ResponseWriter, r *http.Request) {
		records, failures, err := store.ListFiltered(usage.ListOptions{})
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("read store: %v", err))
			return
		}
		writeJSON(w, http.StatusOK, buildSummary(records, failures, opts, store.Root))
	})
	mux.HandleFunc("/api/series", func(w http.ResponseWriter, r *http.Request) {
		metrics, err := requestedMetrics(r.URL.Query().Get("metrics"), r.URL.Query().Get("metric"))
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		since, err := parseSince(r.URL.Query().Get("since"), opts)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		records, _, err := store.ListFiltered(usage.ListOptions{Since: since})
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("read store: %v", err))
			return
		}
		writeJSON(w, http.StatusOK, buildSeries(records, metrics, since, opts))
	})
	return mux
}

func normalizeOptions(opts Options) Options {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	if opts.Since <= 0 {
		opts.Since = DefaultSince
	}
	return opts
}

func dashboardPage(staticDir string) ([]byte, error) {
	if strings.TrimSpace(staticDir) == "" {
		return dashboardHTML, nil
	}
	page, err := os.ReadFile(filepath.Join(staticDir, "dashboard.html"))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", filepath.Join(staticDir, "dashboard.html"), err)
	}
	return page, nil
}

// providerSummary is one provider's newest snapshot plus its volume.
type providerSummary struct {
	Provider   string       `json:"provider"`
	Latest     recentRecord `json:"latest"`
	AgeSeconds float64      `json:"age_seconds"`
	Count      int          `json:"count"`
}

// recentRecord is a stored record with the file it came from.
type recentRecord struct {
	usage.Record
	Path string `json:"path"`
}

type failureJSON struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

type summaryResponse struct {
	Schema        string            `json:"schema"`
	GeneratedAt   time.Time         `json:"generated_at"`
	Root          string            `json:"root"`
	Range         string            `json:"range"`
	RangeSeconds  float64           `json:"range_seconds"`
	Providers     []providerSummary `json:"providers"`
	Recent        []recentRecord    `json:"recent"`
	Metrics       []MetricInfo      `json:"metrics"`
	Errors        []failureJSON     `json:"errors"`
	SnapshotCount int               `json:"snapshot_count"`
}

// buildSummary reduces the stored records into the dashboard's opening payload.
func buildSummary(records []usage.StoredRecord, failures []usage.FileFailure, opts Options, root string) summaryResponse {
	now := opts.now()
	out := summaryResponse{
		Schema:        SchemaID,
		GeneratedAt:   now.UTC(),
		Root:          root,
		Range:         formatRange(opts.Since),
		RangeSeconds:  opts.Since.Seconds(),
		Providers:     []providerSummary{},
		Recent:        []recentRecord{},
		Metrics:       Metrics(records),
		Errors:        []failureJSON{},
		SnapshotCount: len(records),
	}

	// Records are newest first, so the first record of a provider is its latest.
	index := map[usage.ProviderID]int{}
	for _, item := range records {
		position, ok := index[item.Record.Provider]
		if !ok {
			index[item.Record.Provider] = len(out.Providers)
			out.Providers = append(out.Providers, providerSummary{
				Provider:   string(item.Record.Provider),
				Latest:     recentRecord{Record: item.Record, Path: item.Path},
				AgeSeconds: ageSeconds(now, item.Record.TS),
			})
			position = len(out.Providers) - 1
		}
		out.Providers[position].Count++
		if len(out.Recent) < RecentLimit {
			out.Recent = append(out.Recent, recentRecord{Record: item.Record, Path: item.Path})
		}
	}
	sort.Slice(out.Providers, func(i, j int) bool { return out.Providers[i].Provider < out.Providers[j].Provider })

	for _, failure := range failures {
		out.Errors = append(out.Errors, failureJSON{Path: failure.Path, Error: failure.Err.Error()})
	}
	return out
}

// metricBlock is one metric's series inside a /api/series response.
type metricBlock struct {
	Name   string   `json:"name"`
	Unit   string   `json:"unit,omitempty"`
	Kind   string   `json:"kind"`
	Step   bool     `json:"step"`
	Series []Series `json:"series"`
}

type seriesResponse struct {
	Schema  string        `json:"schema"`
	Range   string        `json:"range"`
	Since   time.Time     `json:"since"`
	Metrics []metricBlock `json:"metrics"`
}

func buildSeries(records []usage.StoredRecord, metrics []string, since time.Time, opts Options) seriesResponse {
	out := seriesResponse{
		Schema:  SchemaID,
		Range:   formatRange(opts.Since),
		Since:   since.UTC(),
		Metrics: make([]metricBlock, 0, len(metrics)),
	}
	for _, name := range metrics {
		kind, unit, _ := metricShape(name)
		out.Metrics = append(out.Metrics, metricBlock{
			Name:   name,
			Unit:   unit,
			Kind:   kind,
			Step:   Step(name),
			Series: BuildSeries(records, name, since),
		})
	}
	return out
}

// requestedMetrics resolves the metric query parameters: `metric` names one,
// `metrics` names several, and neither means the two derived metrics.
func requestedMetrics(multi, single string) ([]string, error) {
	var names []string
	switch {
	case strings.TrimSpace(single) != "":
		names = []string{single}
	case strings.TrimSpace(multi) != "":
		names = strings.Split(multi, ",")
	default:
		return append([]string{}, DefaultMetrics...), nil
	}

	seen := map[string]bool{}
	out := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		if strings.ContainsAny(name, " \t") {
			return nil, fmt.Errorf("invalid metric %q", name)
		}
		seen[name] = true
		out = append(out, name)
	}
	if len(out) == 0 {
		return nil, errors.New("no metric requested")
	}
	if len(out) > 16 {
		return nil, fmt.Errorf("too many metrics requested: %d", len(out))
	}
	return out, nil
}

// parseSince resolves the `since` query parameter: a duration such as 24h or
// 7d, `all`, or an RFC3339 instant. Empty uses the server's default range.
func parseSince(value string, opts Options) (time.Time, error) {
	now := opts.now()
	if strings.TrimSpace(value) == "" {
		return now.Add(-opts.Since), nil
	}
	if strings.EqualFold(strings.TrimSpace(value), "all") {
		return time.Time{}, nil
	}
	if at, err := time.Parse(time.RFC3339, value); err == nil {
		return at, nil
	}
	rangeDuration, err := ParseRange(value)
	if err != nil {
		return time.Time{}, err
	}
	return now.Add(-rangeDuration), nil
}

// ParseRange parses a chart range: a duration such as 90m, 24h, 7d or 2w.
func ParseRange(value string) (time.Duration, error) {
	text := strings.TrimSpace(value)
	if text == "" {
		return 0, errors.New("expected a duration like 24h or 7d")
	}
	switch text[len(text)-1] {
	case 'd', 'w':
		count, err := strconv.ParseFloat(text[:len(text)-1], 64)
		if err != nil || count <= 0 {
			return 0, fmt.Errorf("expected a duration like 24h or 7d, got %q", text)
		}
		hours := 24 * count
		if text[len(text)-1] == 'w' {
			hours *= 7
		}
		return time.Duration(hours * float64(time.Hour)), nil
	default:
		parsed, err := time.ParseDuration(text)
		if err != nil || parsed <= 0 {
			return 0, fmt.Errorf("expected a duration like 24h or 7d, got %q", text)
		}
		return parsed, nil
	}
}

// formatRange renders a range the way it is typed: 7d, 24h, 90m.
func formatRange(d time.Duration) string {
	if d <= 0 {
		return "all"
	}
	if d%(24*time.Hour) == 0 {
		return fmt.Sprintf("%dd", int(d/(24*time.Hour)))
	}
	return d.String()
}

func ageSeconds(now, ts time.Time) float64 {
	seconds := now.Sub(ts).Seconds()
	if seconds < 0 {
		return 0
	}
	return seconds
}

// printStartup reports the bound URL and what the dashboard is reading.
func printStartup(opts Options, store *usage.Store, records []usage.StoredRecord, failures, port int) {
	if opts.PortExplicit && opts.Port > 0 && port != opts.Port {
		fmt.Fprintf(opts.Stderr, "warning: port %d is busy, using %d\n", opts.Port, port)
	}
	url := urlFor(port)
	if opts.JSON {
		payload := struct {
			Schema    string `json:"schema"`
			URL       string `json:"url"`
			Port      int    `json:"port"`
			Root      string `json:"root"`
			Providers int    `json:"providers"`
			Snapshots int    `json:"snapshots"`
			Range     string `json:"range"`
		}{
			Schema:    SchemaID,
			URL:       url,
			Port:      port,
			Root:      store.Root,
			Providers: providerCount(records),
			Snapshots: len(records),
			Range:     formatRange(opts.Since),
		}
		line, err := json.Marshal(payload)
		if err == nil {
			fmt.Fprintln(opts.Stdout, string(line))
			return
		}
	}

	fmt.Fprintf(opts.Stdout, "serving %s\n", url)
	summary := "store   " + displayPath(store.Root, homeDir())
	summary += fmt.Sprintf("  (read-only, %d provider%s, %d snapshot%s",
		providerCount(records), plural(providerCount(records)), len(records), plural(len(records)))
	if failures > 0 {
		summary += fmt.Sprintf(", %d unreadable", failures)
	}
	fmt.Fprintln(opts.Stdout, summary+")")
	if newest, ok := newestRecord(records); ok {
		fmt.Fprintf(opts.Stdout, "newest  %s (%s ago)\n", newest.UTC().Format(time.RFC3339), shortAge(opts.now(), newest))
	}
	fmt.Fprintf(opts.Stdout, "range   %s\n", formatRange(opts.Since))
}

func providerCount(records []usage.StoredRecord) int {
	seen := map[usage.ProviderID]bool{}
	for _, item := range records {
		seen[item.Record.Provider] = true
	}
	return len(seen)
}

func newestRecord(records []usage.StoredRecord) (time.Time, bool) {
	if len(records) == 0 {
		return time.Time{}, false
	}
	newest := records[0].Record.TS
	for _, item := range records[1:] {
		if item.Record.TS.After(newest) {
			newest = item.Record.TS
		}
	}
	return newest, true
}

func plural(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// shortAge renders an age the way the dashboard header does: 12s, 3m, 4h, 2d.
func shortAge(now, ts time.Time) string {
	seconds := int(ageSeconds(now, ts))
	switch {
	case seconds < 60:
		return fmt.Sprintf("%ds", seconds)
	case seconds < 3600:
		return fmt.Sprintf("%dm", seconds/60)
	case seconds < 86400:
		return fmt.Sprintf("%dh", seconds/3600)
	default:
		return fmt.Sprintf("%dd", seconds/86400)
	}
}

// listenLocalPreferred binds 127.0.0.1 at start, trying the next port when it is
// busy, for up to viewPortAttempts ports. Zero or negative binds any free port.
func listenLocalPreferred(start int) (net.Listener, error) {
	return listenRange(start, viewPortAttempts)
}

func listenRange(start, attempts int) (net.Listener, error) {
	if start <= 0 {
		return net.Listen("tcp", "127.0.0.1:0")
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", start+i))
		if err == nil {
			return listener, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("no free port in %d..%d: %w", start, start+attempts-1, lastErr)
}

func listenerPort(listener net.Listener) int {
	if addr, ok := listener.Addr().(*net.TCPAddr); ok {
		return addr.Port
	}
	return 0
}

func urlFor(port int) string {
	return fmt.Sprintf("http://127.0.0.1:%d", port)
}

// openLocalURL opens url in the platform browser, best effort.
func openLocalURL(url string) {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", url)
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		command = exec.Command("xdg-open", url)
	}
	_ = command.Start()
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// displayPath shortens a path under home to ~, for readable output.
func displayPath(path, home string) string {
	if home == "" {
		return path
	}
	if path == home {
		return "~"
	}
	if strings.HasPrefix(path, home+string(filepath.Separator)) {
		return "~" + path[len(home):]
	}
	return path
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("encode response: %v", err))
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = w.Write(append(body, '\n'))
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"schema": SchemaID, "error": message})
}

func writeText(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}
