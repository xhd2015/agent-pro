package idle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
	"unicode/utf8"
)

// Event names written to idle.jsonl.
const (
	EventArmed    = "idle.armed"
	EventTick     = "idle.tick"
	EventReset    = "idle.reset"
	EventSoftExit = "idle.soft_exit"
	EventShutdown = "idle.shutdown"
)

// Reset reasons on idle.reset.
const (
	ResetSnapshotErr = "snapshot_err"
	ResetChanged     = "changed"
	ResetNotReady    = "not_ready"
	ResetOccupied    = "occupied"
	ResetQueue       = "queue"
	ResetFinalOccupy = "final_occupy"
)

// DefaultSnapTailRunes is how many trailing runes to keep on soft_exit / reset=changed.
const DefaultSnapTailRunes = 200

// Event is one idle.jsonl line (compact JSON object).
type Event struct {
	TS        time.Time `json:"ts"`
	Event     string    `json:"event"`
	SessionID string    `json:"session_id,omitempty"`
	Timeout   string    `json:"timeout,omitempty"`
	Grace     string    `json:"grace,omitempty"`
	Hits      int       `json:"hits,omitempty"`
	HitsBefore int      `json:"hits_before,omitempty"`
	IdleSince string    `json:"idle_since,omitempty"`
	Age       string    `json:"age,omitempty"`
	Ready     *bool     `json:"ready,omitempty"`
	ReadyState string   `json:"ready_state,omitempty"`
	Occupy    string    `json:"occupy,omitempty"`
	Queue     *int      `json:"queue,omitempty"`
	Reason    string    `json:"reason,omitempty"`
	SnapHash  string    `json:"snap_hash,omitempty"`
	SnapLen   int       `json:"snap_len,omitempty"`
	SnapTail  string    `json:"snap_tail,omitempty"`
}

// SnapMeta returns sha256 hex (16 chars) and byte length for a resting snapshot.
func SnapMeta(snap string) (hash string, length int) {
	sum := sha256.Sum256([]byte(snap))
	return hex.EncodeToString(sum[:8]), len(snap)
}

// SnapTail returns the last n runes of snap (n<=0 → DefaultSnapTailRunes).
func SnapTail(snap string, n int) string {
	if n <= 0 {
		n = DefaultSnapTailRunes
	}
	if snap == "" {
		return ""
	}
	if utf8.RuneCountInString(snap) <= n {
		return snap
	}
	runes := []rune(snap)
	return string(runes[len(runes)-n:])
}

// JSONL appends Events as one JSON object per line. Safe for concurrent use.
type JSONL struct {
	Path string

	mu sync.Mutex
}

// Log appends e as a single JSON line. Creates the parent dir if needed.
// Failures are ignored (idle exit must not fail open because of logging).
func (j *JSONL) Log(e Event) {
	if j == nil || j.Path == "" {
		return
	}
	if e.TS.IsZero() {
		e.TS = time.Now().UTC()
	} else {
		e.TS = e.TS.UTC()
	}
	body, err := json.Marshal(e)
	if err != nil {
		return
	}
	body = append(body, '\n')

	j.mu.Lock()
	defer j.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(j.Path), 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(j.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	_, _ = f.Write(body)
	_ = f.Close()
}
