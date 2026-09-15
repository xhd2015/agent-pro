package view

import (
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/agent/usage"
)

func stored(provider usage.ProviderID, at time.Time, values map[string]float64, sessions int) usage.StoredRecord {
	return usage.StoredRecord{Record: testRecord(provider, at, values, sessions), Path: "/store/" + string(provider)}
}

func TestBuildSeriesWindowAndOrder(t *testing.T) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	// Newest first, the order Store.List returns.
	records := []usage.StoredRecord{
		stored(usage.Grok, now.Add(-2*time.Hour), map[string]float64{"used_percent": 30}, 3),
		stored(usage.Grok, now.Add(-26*time.Hour), map[string]float64{"used_percent": 20}, 2),
		stored(usage.Grok, now.Add(-50*time.Hour), map[string]float64{"used_percent": 10}, 1),
	}

	all := BuildSeries(records, "used_percent", time.Time{})
	if len(all) != 1 {
		t.Fatalf("got %d series, want 1", len(all))
	}
	if got := len(all[0].Points); got != 3 {
		t.Fatalf("unbounded series has %d points, want 3", got)
	}
	if all[0].Points[0].Value != 10 || all[0].Points[2].Value != 30 {
		t.Fatalf("series is not ascending: %+v", all[0].Points)
	}
	if all[0].Latest != 30 || all[0].Min != 10 || all[0].Max != 30 {
		t.Fatalf("series summary = latest %v min %v max %v, want 30/10/30",
			all[0].Latest, all[0].Min, all[0].Max)
	}

	windowed := BuildSeries(records, "used_percent", now.Add(-24*time.Hour))
	if len(windowed) != 1 || len(windowed[0].Points) != 1 {
		t.Fatalf("24h window = %+v, want one point", windowed)
	}
	if windowed[0].Points[0].Value != 30 {
		t.Fatalf("24h window value = %v, want the newest 30", windowed[0].Points[0].Value)
	}
}

func TestBuildSeriesPrimaryPercentAlias(t *testing.T) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	records := []usage.StoredRecord{
		stored(usage.CommandCode, now, map[string]float64{"usage_percent": 29, "cost_usd": 12.25}, 190),
		stored(usage.Codex, now, map[string]float64{"used_percent": 61}, 180),
		stored(usage.Grok, now, map[string]float64{"used_percent": 42}, 178),
	}

	series := BuildSeries(records, MetricPrimaryPercent, time.Time{})
	if len(series) != 3 {
		t.Fatalf("primary_percent covers %d providers, want 3", len(series))
	}
	want := map[string]float64{"codex": 61, "commandcode": 29, "grok": 42}
	for _, item := range series {
		if got := item.Latest; got != want[item.Provider] {
			t.Fatalf("%s primary_percent = %v, want %v", item.Provider, got, want[item.Provider])
		}
	}

	// A provider whose percentage key is missing contributes no point.
	for _, metric := range []string{"used_percent", "usage_percent"} {
		byProvider := map[string]bool{}
		for _, item := range BuildSeries(records, metric, time.Time{}) {
			byProvider[item.Provider] = true
		}
		if metric == "used_percent" && (byProvider["commandcode"] || !byProvider["grok"]) {
			t.Fatalf("used_percent providers = %v, want grok and codex only", byProvider)
		}
		if metric == "usage_percent" && (byProvider["grok"] || !byProvider["commandcode"]) {
			t.Fatalf("usage_percent providers = %v, want commandcode only", byProvider)
		}
	}
}

func TestBuildSeriesSessionsTotalSkipsUnreadable(t *testing.T) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	unreadable := stored(usage.Grok, now, map[string]float64{"used_percent": 10}, 0)
	unreadable.Record.Sessions.Error = "read sessions: permission denied"
	failedUsage := stored(usage.Codex, now, nil, 12)
	failedUsage.Record.Usage.OK = false
	failedUsage.Record.Usage.Error = "fetch: no credentials"

	series := BuildSeries([]usage.StoredRecord{unreadable, failedUsage}, MetricSessionsTotal, time.Time{})
	if len(series) != 1 || series[0].Provider != "codex" {
		t.Fatalf("sessions_total series = %+v, want codex only", series)
	}
	if series[0].Latest != 12 {
		t.Fatalf("codex sessions_total = %v, want 12 (a failed usage fetch keeps its counts)", series[0].Latest)
	}

	// A failed usage fetch charts no percentage.
	if got := BuildSeries([]usage.StoredRecord{failedUsage}, MetricPrimaryPercent, time.Time{}); len(got) != 0 {
		t.Fatalf("primary_percent picked up a failed fetch: %+v", got)
	}
}

func TestBuildSeriesNewestWinsWithinSecond(t *testing.T) {
	at := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	records := []usage.StoredRecord{
		stored(usage.Grok, at, map[string]float64{"used_percent": 30}, 1),
		stored(usage.Grok, at, map[string]float64{"used_percent": 20}, 1),
	}
	series := BuildSeries(records, "used_percent", time.Time{})
	if len(series) != 1 || len(series[0].Points) != 1 {
		t.Fatalf("same-second points = %+v, want one point", series)
	}
	if series[0].Latest != 30 {
		t.Fatalf("same-second value = %v, want the newest 30", series[0].Latest)
	}
}

func TestMetricsDiscovery(t *testing.T) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)
	records := []usage.StoredRecord{
		stored(usage.Grok, now, map[string]float64{"used_percent": 42, "reset_at": 1788220800}, 178),
		stored(usage.CommandCode, now, map[string]float64{"usage_percent": 29, "cost_usd": 12.25}, 190),
	}

	metrics := Metrics(records)
	byName := map[string]MetricInfo{}
	for _, metric := range metrics {
		byName[metric.Name] = metric
	}
	if metrics[0].Name != MetricPrimaryPercent || metrics[1].Name != MetricSessionsTotal {
		t.Fatalf("metrics start with %s, %s, want the derived pair first", metrics[0].Name, metrics[1].Name)
	}
	if got := byName[MetricPrimaryPercent]; got.Count != 2 || len(got.Providers) != 2 {
		t.Fatalf("primary_percent = %+v, want both providers", got)
	}
	if got := byName["used_percent"]; len(got.Providers) != 1 || got.Providers[0] != "grok" {
		t.Fatalf("used_percent = %+v, want grok only", got)
	}
	if got := byName[MetricSessionsTotal]; got.Kind != "counter" || !Step(MetricSessionsTotal) {
		t.Fatalf("sessions_total = %+v, want a counter drawn as steps", got)
	}
	reset, ok := byName["reset_at"]
	if !ok {
		t.Fatal("reset_at is missing from the metric list")
	}
	if reset.Chartable || reset.Kind != "time" {
		t.Fatalf("reset_at = %+v, want a non-chartable timestamp", reset)
	}
	if got := byName["used_percent"].Unit; got != "%" {
		t.Fatalf("used_percent unit = %q, want %%", got)
	}
}
