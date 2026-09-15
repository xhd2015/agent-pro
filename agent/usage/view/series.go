// Package view serves a read-only dashboard of the usage snapshots written by
// `agent-pro usage collect`. It reads the store, never writes to it.
package view

import (
	"sort"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/agent/usage"
)

// Metric names the dashboard adds to the keys stored in a snapshot.
const (
	// MetricPrimaryPercent overlays each provider's own quota percentage, which
	// every provider names differently.
	MetricPrimaryPercent = "primary_percent"
	// MetricSessionsTotal is the count of sessions found on disk per provider.
	MetricSessionsTotal = "sessions_total"
)

// DefaultMetrics are charted when a request names no metric.
var DefaultMetrics = []string{MetricPrimaryPercent, MetricSessionsTotal}

// timeMetricKeys hold a unix timestamp rather than a magnitude, so they are
// shown as dates and left out of the chart menu.
var timeMetricKeys = map[string]bool{
	"reset_at":        true,
	"period_start":    true,
	"period_end":      true,
	"weekly_reset_at": true,
}

// primaryPercentKey maps a provider to the snapshot key holding its quota
// percentage.
var primaryPercentKey = map[usage.ProviderID]string{
	usage.Grok:        "used_percent",
	usage.Codex:       "used_percent",
	usage.CommandCode: "usage_percent",
}

// MetricInfo describes one metric the dashboard can chart.
type MetricInfo struct {
	Name string `json:"name"`
	// Kind is percent, counter or number; KindTime marks a timestamp.
	Kind string `json:"kind"`
	Unit string `json:"unit,omitempty"`
	// Providers are the providers that reported this metric, sorted.
	Providers []string `json:"providers"`
	// Count is how many stored records carry a value for it.
	Count int `json:"count"`
	// Chartable is false for values that are timestamps, not magnitudes.
	Chartable bool `json:"chartable"`
}

// Point is one charted value.
type Point struct {
	TS    time.Time `json:"ts"`
	Value float64   `json:"value"`
}

// Series is one provider's points for one metric, ascending by time.
type Series struct {
	Provider string  `json:"provider"`
	Points   []Point `json:"points"`
	Latest   float64 `json:"latest"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
}

// Metrics lists every chartable metric found in the records, the two derived
// ones first, then the raw snapshot keys alphabetically.
func Metrics(records []usage.StoredRecord) []MetricInfo {
	seen := map[string]map[usage.ProviderID]bool{}
	counts := map[string]int{}

	for _, item := range records {
		rec := item.Record
		for key := range rec.Usage.Values {
			seen[key] = addProvider(seen[key], rec.Provider)
			counts[key]++
		}
		if rec.Sessions.Error == "" {
			seen[MetricSessionsTotal] = addProvider(seen[MetricSessionsTotal], rec.Provider)
			counts[MetricSessionsTotal]++
		}
		if key, ok := primaryPercentKey[rec.Provider]; ok {
			if _, has := rec.Usage.Values[key]; has {
				seen[MetricPrimaryPercent] = addProvider(seen[MetricPrimaryPercent], rec.Provider)
				counts[MetricPrimaryPercent]++
			}
		}
	}

	out := make([]MetricInfo, 0, len(seen))
	for name, providers := range seen {
		kind, unit, chartable := metricShape(name)
		out = append(out, MetricInfo{
			Name:      name,
			Kind:      kind,
			Unit:      unit,
			Providers: sortedProviders(providers),
			Count:     counts[name],
			Chartable: chartable,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		left, right := metricRank(out[i].Name), metricRank(out[j].Name)
		if left != right {
			return left < right
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// metricRank orders the derived metrics before the raw snapshot keys.
func metricRank(name string) int {
	switch name {
	case MetricPrimaryPercent:
		return 0
	case MetricSessionsTotal:
		return 1
	default:
		return 2
	}
}

// metricShape classifies a metric for axis labels and step rendering.
func metricShape(name string) (kind, unit string, chartable bool) {
	if timeMetricKeys[name] {
		return "time", "", false
	}
	switch {
	case strings.HasSuffix(name, "_percent"):
		return "percent", "%", true
	case strings.HasSuffix(name, "_total"):
		return "counter", "", true
	default:
		return "number", "", true
	}
}

// Chartable reports whether a metric name can be plotted.
func Chartable(name string) bool {
	_, _, ok := metricShape(name)
	return ok
}

// Step reports whether a metric accumulates, so it is drawn as a step line.
func Step(name string) bool {
	kind, _, _ := metricShape(name)
	return kind == "counter"
}

func addProvider(set map[usage.ProviderID]bool, id usage.ProviderID) map[usage.ProviderID]bool {
	if set == nil {
		set = map[usage.ProviderID]bool{}
	}
	set[id] = true
	return set
}

func sortedProviders(set map[usage.ProviderID]bool) []string {
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, string(id))
	}
	sort.Strings(out)
	return out
}

// BuildSeries pivots stored records into per-provider series for one metric.
// Records arrive newest first (store order) and each series is returned
// ascending by time, so a chart can draw it directly.
func BuildSeries(records []usage.StoredRecord, metric string, since time.Time) []Series {
	byProvider := map[usage.ProviderID]map[int64]float64{}
	for _, item := range records {
		rec := item.Record
		if !since.IsZero() && rec.TS.Before(since) {
			continue
		}
		value, ok := metricValue(rec, metric)
		if !ok {
			continue
		}
		points := byProvider[rec.Provider]
		if points == nil {
			points = map[int64]float64{}
			byProvider[rec.Provider] = points
		}
		// Records are newest first, so the first value seen for a second is the
		// newest snapshot of that second: a second run never rewrites an older one.
		if _, seen := points[rec.TS.Unix()]; seen {
			continue
		}
		points[rec.TS.Unix()] = value
	}

	out := make([]Series, 0, len(byProvider))
	for provider, points := range byProvider {
		seconds := make([]int64, 0, len(points))
		for second := range points {
			seconds = append(seconds, second)
		}
		sort.Slice(seconds, func(i, j int) bool { return seconds[i] < seconds[j] })

		series := Series{
			Provider: string(provider),
			Points:   make([]Point, 0, len(seconds)),
			Min:      points[seconds[0]],
			Max:      points[seconds[0]],
		}
		for _, second := range seconds {
			value := points[second]
			series.Points = append(series.Points, Point{
				TS:    time.Unix(second, 0).UTC(),
				Value: value,
			})
			if value < series.Min {
				series.Min = value
			}
			if value > series.Max {
				series.Max = value
			}
		}
		series.Latest = series.Points[len(series.Points)-1].Value
		out = append(out, series)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Provider < out[j].Provider })
	return out
}

// metricValue reads one charted value out of a record. A record whose usage
// fetch failed has no usage values, but its session count still charts; a
// record whose session tree was unreadable charts no session count.
func metricValue(rec usage.Record, metric string) (float64, bool) {
	switch metric {
	case MetricPrimaryPercent:
		key, ok := primaryPercentKey[rec.Provider]
		if !ok {
			return 0, false
		}
		value, ok := rec.Usage.Values[key]
		return value, ok
	case MetricSessionsTotal:
		if rec.Sessions.Error != "" {
			return 0, false
		}
		return float64(rec.Sessions.Total), true
	default:
		value, ok := rec.Usage.Values[metric]
		return value, ok
	}
}
