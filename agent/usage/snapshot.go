package usage

import (
	"encoding/json"
	"time"
)

// SchemaID is the on-disk schema for one usage snapshot record.
const SchemaID = "agent-pro/usage-snapshot/v1"

// Record is one collected snapshot for a single provider: the account usage
// read over that provider's HTTP API plus the counts of the sessions the
// provider has on disk. Records are appended one per line to
// $AGENT_PRO_HOME/usages/<provider>/<date>/<time>-snapshot.jsonl.
type Record struct {
	Schema   string        `json:"schema"`
	TS       time.Time     `json:"ts"`
	Provider ProviderID    `json:"provider"`
	Usage    UsageBlock    `json:"usage"`
	Sessions SessionsBlock `json:"sessions"`
}

// UsageBlock is one provider's account usage as returned by its API.
// Values holds the numbers worth charting; Display keeps the provider's own
// rendered strings (plan name, email, reset text) that do not belong in Values.
type UsageBlock struct {
	OK         bool               `json:"ok"`
	Endpoint   string             `json:"endpoint,omitempty"`
	DurationMS int64              `json:"duration_ms"`
	Values     map[string]float64 `json:"values,omitempty"`
	Display    map[string]string  `json:"display,omitempty"`
	Error      string             `json:"error,omitempty"`
}

// SessionsBlock counts the provider's sessions as found on disk. Oldest is the
// earliest session start and Newest the latest recorded activity; both are
// omitted when the provider has no readable sessions.
type SessionsBlock struct {
	Total  int        `json:"total"`
	Oldest *time.Time `json:"oldest,omitempty"`
	Newest *time.Time `json:"newest,omitempty"`
	Error  string     `json:"error,omitempty"`
}

// MarshalJSONLine returns the record as compact single-line JSON, the form
// written to the store and printed by --json.
func (r Record) MarshalJSONLine() ([]byte, error) {
	return json.Marshal(r)
}

// Headline renders the short human usage summary ("usage" display key).
func (r Record) Headline() string {
	if !r.Usage.OK {
		return "(failed)"
	}
	if headline := r.Usage.Display["usage"]; headline != "" {
		return headline
	}
	return "ok"
}
