package knowledgesink

import (
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/agent/grok/sessions"
)

// Button / status states for callers (Marcus UI, CLI).
const (
	StateUnavailable = "unavailable"
	StateReady       = "ready"
	StateSunk        = "sunk"
	StateBehind      = "behind"
	StateRunning     = "running"
	StateFailed      = "failed"
)

// StatusView is the sink button / status snapshot.
type StatusView struct {
	State    string `json:"state"` // unavailable|ready|sunk|behind|running|failed
	Label    string `json:"label"`
	Enabled  bool   `json:"enabled"`
	Help     string `json:"help,omitempty"`
	Error    string `json:"error,omitempty"`
	NewCount int    `json:"new_count,omitempty"`
}

func BuildStatus(manifest *Manifest, tip time.Time, total int, grokOK bool, grokHelp string) *StatusView {
	// Manifest running is SSOT — show Sinking… even if runner probe fails mid-run.
	if manifest != nil && strings.EqualFold(manifest.Status, statusRunning) {
		return &StatusView{
			State:   StateRunning,
			Label:   "Sinking…",
			Enabled: false,
			Help:    "Knowledge sink is running",
		}
	}
	if !grokOK {
		return &StatusView{
			State:   StateUnavailable,
			Label:   "Sink Knowledge",
			Enabled: false,
			Help:    firstNonEmpty(grokHelp, "Needs a grok or codex agent session (runner Session ID)"),
		}
	}
	// neverSunk = no history and no checked/sunk tip. History without a tip
	// cursor still counts as sunk for auto-pick / UI; use LastSinkAt as fallback
	// so tip-after can detect real new work. Sinkability uses checked (falls
	// back to sunk for older manifests).
	var lastMax time.Time
	if manifest != nil {
		if t, err := ParseTime(CheckedCursor(manifest)); err == nil {
			lastMax = t
		} else if HasSinkHistory(manifest) {
			if t, err := ParseTime(manifest.LastSinkAt); err == nil {
				lastMax = t
			}
		}
	}
	neverSunk := manifest == nil || (!HasSinkHistory(manifest) && CheckedCursor(manifest) == "")
	if neverSunk {
		label := "Sink Knowledge"
		enabled := total > 0
		help := "Propose knowledge from the runner session into the knowledge base"
		if !enabled {
			help = "No session messages to sink yet"
		}
		out := &StatusView{State: StateReady, Label: label, Enabled: enabled, Help: help}
		if manifest != nil && strings.EqualFold(manifest.Status, statusFailed) {
			out.State = StateFailed
			out.Error = manifest.Error
			out.Enabled = total > 0
			out.Label = "Sink Knowledge"
			out.Help = firstNonEmpty(manifest.Error, out.Help)
		}
		return out
	}
	// History exists but neither cursor nor last_sink_at parsed → do not claim behind.
	if lastMax.IsZero() || !TipAfterMax(tip, lastMax) {
		return &StatusView{
			State:   StateSunk,
			Label:   "Sinked",
			Enabled: false,
			Help:    "Already sunk through the latest Grok message",
		}
	}
	out := &StatusView{
		State:   StateBehind,
		Label:   "Sink more",
		Enabled: true,
		Help:    "New Grok messages since last sink",
	}
	if manifest != nil && strings.EqualFold(manifest.Status, statusFailed) {
		out.State = StateFailed
		out.Error = manifest.Error
		out.Help = firstNonEmpty(manifest.Error, out.Help)
	}
	return out
}

func NewestMessageTime(msgs []sessions.ChatMessage) time.Time {
	var tip time.Time
	for _, m := range msgs {
		if m.Timestamp.IsZero() {
			continue
		}
		if tip.IsZero() || m.Timestamp.After(tip) {
			tip = m.Timestamp
		}
	}
	return tip
}

func FilterMessagesAfter(msgs []sessions.ChatMessage, after time.Time) []sessions.ChatMessage {
	if after.IsZero() {
		return msgs
	}
	out := make([]sessions.ChatMessage, 0, len(msgs))
	for _, m := range msgs {
		if m.Timestamp.IsZero() {
			out = append(out, m)
			continue
		}
		if m.Timestamp.After(after) {
			out = append(out, m)
		}
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
