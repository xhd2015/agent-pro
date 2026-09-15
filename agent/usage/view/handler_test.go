package view

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/usage"
)

// getJSON decodes one dashboard API response, failing on a non-200 status.
func getJSON(t *testing.T, baseURL, path string, target any) {
	t.Helper()
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s = %d, want 200", path, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}

type summaryPayload struct {
	Schema        string  `json:"schema"`
	Root          string  `json:"root"`
	Range         string  `json:"range"`
	RangeSeconds  float64 `json:"range_seconds"`
	SnapshotCount int     `json:"snapshot_count"`
	Providers     []struct {
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
			} `json:"usage"`
			Sessions struct {
				Total int `json:"total"`
			} `json:"sessions"`
			Path string `json:"path"`
		} `json:"latest"`
	} `json:"providers"`
	Recent []struct {
		TS       time.Time `json:"ts"`
		Provider string    `json:"provider"`
		Path     string    `json:"path"`
	} `json:"recent"`
	Metrics []MetricInfo `json:"metrics"`
	Errors  []struct {
		Path  string `json:"path"`
		Error string `json:"error"`
	} `json:"errors"`
}

type seriesPayload struct {
	Schema  string `json:"schema"`
	Range   string `json:"range"`
	Metrics []struct {
		Name   string   `json:"name"`
		Unit   string   `json:"unit"`
		Kind   string   `json:"kind"`
		Step   bool     `json:"step"`
		Series []Series `json:"series"`
	} `json:"metrics"`
}

func newServer(t *testing.T, store *usage.Store, opts Options) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(Handler(store, opts))
	t.Cleanup(server.Close)
	return server
}

func TestIndexServesDashboard(t *testing.T) {
	store := testStore(t, testRecord(usage.Grok, time.Now().Add(-time.Hour), map[string]float64{"used_percent": 42}, 7))
	server := newServer(t, store, Options{})

	body := getBody(t, server.URL+"/")
	markers := []string{
		"<title>agent-pro usage</title>", "id=\"groups\"", "/api/summary", "/api/series",
		// Without a viewport meta, mobile browsers use a ~980px layout viewport;
		// minmax(0, 1fr) keeps the main column at the viewport width on narrow screens.
		"name=\"viewport\"", "grid-template-columns: minmax(0, 1fr)",
		// The dashboard draws one group per provider in this order.
		"var PROVIDER_ORDER = [\"codex\", \"grok\", \"commandcode\"]",
		// Charts come from a pinned d3 build; the exact URL and hash are the guard
		// against an accidental move to a floating version or a stale digest.
		"https://cdn.jsdelivr.net/npm/d3@7.9.0/dist/d3.min.js",
		"sha384-CjloA8y00+1SDAUkjs099PVfnY2KmDC2BZnws9kh8D/lX1s46w6EPhpXdqMfjK6i",
	}
	for _, marker := range markers {
		if !strings.Contains(body, marker) {
			t.Fatalf("dashboard page is missing %q", marker)
		}
	}

	// Anything outside the dashboard root is not served.
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(server.URL + "/nope")
	if err != nil {
		t.Fatalf("GET /nope: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("GET /nope = %d, want 404", resp.StatusCode)
	}
}

func TestSummaryLatestPerProviderAndRecent(t *testing.T) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	store := testStore(t,
		testRecord(usage.Grok, now.Add(-3*time.Hour), map[string]float64{"used_percent": 30}, 100),
		testRecord(usage.Grok, now.Add(-time.Hour), map[string]float64{"used_percent": 42}, 178),
		testRecord(usage.Codex, now.Add(-2*time.Hour), map[string]float64{"used_percent": 61}, 92),
		testRecord(usage.CommandCode, now.Add(-30*time.Minute), map[string]float64{"usage_percent": 29}, 41),
	)
	server := newServer(t, store, Options{Since: 7 * 24 * time.Hour, Now: clock})

	var summary summaryPayload
	getJSON(t, server.URL, "/api/summary", &summary)

	if summary.Schema != SchemaID {
		t.Fatalf("schema = %q, want %q", summary.Schema, SchemaID)
	}
	if summary.Range != "7d" || summary.SnapshotCount != 4 {
		t.Fatalf("range = %q snapshots = %d, want 7d and 4", summary.Range, summary.SnapshotCount)
	}
	if summary.Root != store.Root {
		t.Fatalf("root = %q, want %q", summary.Root, store.Root)
	}
	if len(summary.Providers) != 3 {
		t.Fatalf("providers = %d, want 3", len(summary.Providers))
	}
	for _, entry := range summary.Providers {
		switch entry.Provider {
		case "grok":
			if entry.Latest.Usage.Values["used_percent"] != 42 || entry.Count != 2 {
				t.Fatalf("grok = %+v, want the newest 42%% over 2 snapshots", entry)
			}
			if want := 3600.0; entry.AgeSeconds != want {
				t.Fatalf("grok age = %v, want %v", entry.AgeSeconds, want)
			}
			if entry.Latest.Sessions.Total != 178 {
				t.Fatalf("grok sessions = %d, want 178", entry.Latest.Sessions.Total)
			}
			if !strings.HasSuffix(entry.Latest.Path, "-snapshot.jsonl") {
				t.Fatalf("grok latest path = %q, want the snapshot file", entry.Latest.Path)
			}
		case "codex", "commandcode":
		default:
			t.Fatalf("unexpected provider %q", entry.Provider)
		}
	}

	if len(summary.Recent) != 4 {
		t.Fatalf("recent = %d records, want all 4", len(summary.Recent))
	}
	// Newest first, across providers.
	wantOrder := []string{"commandcode", "grok", "codex", "grok"}
	for i, want := range wantOrder {
		if summary.Recent[i].Provider != want {
			t.Fatalf("recent[%d] = %s, want %s", i, summary.Recent[i].Provider, want)
		}
	}
	if len(summary.Metrics) == 0 {
		t.Fatal("summary carries no metrics for the dashboard menu")
	}
}

func TestSeriesEndpointWindowAndUnknownMetric(t *testing.T) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	store := testStore(t,
		testRecord(usage.Grok, now.Add(-2*time.Hour), map[string]float64{"used_percent": 30}, 3),
		testRecord(usage.Grok, now.Add(-50*time.Hour), map[string]float64{"used_percent": 10}, 1),
	)
	server := newServer(t, store, Options{Since: 7 * 24 * time.Hour, Now: clock})

	var windowed seriesPayload
	getJSON(t, server.URL, "/api/series?metric=used_percent&since=24h", &windowed)
	if len(windowed.Metrics) != 1 || windowed.Metrics[0].Name != "used_percent" {
		t.Fatalf("metrics = %+v, want one used_percent block", windowed.Metrics)
	}
	points := windowed.Metrics[0].Series[0].Points
	if len(points) != 1 || points[0].Value != 30 {
		t.Fatalf("24h window points = %+v, want the -2h point only", points)
	}
	if !points[0].TS.Equal(now.Add(-2 * time.Hour)) {
		t.Fatalf("point ts = %s, want %s", points[0].TS, now.Add(-2*time.Hour))
	}

	var all seriesPayload
	getJSON(t, server.URL, "/api/series?metrics=used_percent,sessions_total&since=all", &all)
	if len(all.Metrics) != 2 {
		t.Fatalf("metrics = %d blocks, want 2", len(all.Metrics))
	}
	if got := len(all.Metrics[0].Series[0].Points); got != 2 {
		t.Fatalf("all-time points = %d, want 2", got)
	}

	var unknown seriesPayload
	getJSON(t, server.URL, "/api/series?metric=does_not_exist", &unknown)
	if len(unknown.Metrics) != 1 || len(unknown.Metrics[0].Series) != 0 {
		t.Fatalf("unknown metric = %+v, want an empty series at 200", unknown.Metrics)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(server.URL + "/api/series?since=bogus")
	if err != nil {
		t.Fatalf("GET bad since: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("bad since = %d, want 400", resp.StatusCode)
	}
}

func TestEmptyStoreServesDashboard(t *testing.T) {
	store := usage.NewStore(t.TempDir())
	server := newServer(t, store, Options{})

	body := getBody(t, server.URL+"/")
	if !strings.Contains(body, "id=\"empty\"") {
		t.Fatal("dashboard page has no empty state")
	}

	var summary summaryPayload
	getJSON(t, server.URL, "/api/summary", &summary)
	if len(summary.Providers) != 0 || len(summary.Recent) != 0 || summary.SnapshotCount != 0 {
		t.Fatalf("empty store summary = %+v, want no providers and no records", summary)
	}
	if summary.Errors == nil || summary.Metrics == nil {
		t.Fatalf("empty store summary = %+v, want empty arrays rather than null", summary)
	}
}

func TestUnreadableSnapshotIsSkipped(t *testing.T) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	store := testStore(t, testRecord(usage.Grok, now.Add(-time.Hour), map[string]float64{"used_percent": 42}, 7))
	dir := store.DayDir(usage.Codex, now)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	broken := filepath.Join(dir, "01-02-03"+usage.SnapshotSuffix)
	if err := os.WriteFile(broken, []byte("{not json}\n"), 0o644); err != nil {
		t.Fatalf("write broken snapshot: %v", err)
	}
	server := newServer(t, store, Options{Now: func() time.Time { return now }})

	var summary summaryPayload
	getJSON(t, server.URL, "/api/summary", &summary)
	if len(summary.Providers) != 1 || summary.Providers[0].Provider != "grok" {
		t.Fatalf("providers = %+v, want grok only", summary.Providers)
	}
	if len(summary.Errors) != 1 || summary.Errors[0].Path != broken {
		t.Fatalf("errors = %+v, want the unreadable file reported", summary.Errors)
	}
	if !strings.Contains(summary.Errors[0].Error, "parse") {
		t.Fatalf("error = %q, want the parse failure", summary.Errors[0].Error)
	}
}

func TestStaticDirOverridesEmbeddedPage(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "dashboard.html"), []byte("<html>from disk</html>"), 0o644); err != nil {
		t.Fatalf("write dashboard.html: %v", err)
	}
	store := usage.NewStore(t.TempDir())
	server := newServer(t, store, Options{StaticDir: dir})

	if body := getBody(t, server.URL+"/"); !strings.Contains(body, "from disk") {
		t.Fatalf("page = %q, want the file from --static-dir", body)
	}
}
