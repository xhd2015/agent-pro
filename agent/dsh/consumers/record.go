package consumers

import (
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/xhd2015/agent-pro/agent/dsh/sessions"
)

// record is the per-session cache. It holds the stream cursor plus one small
// state per consumer, which is what lets the next run read only new bytes.
type record struct {
	Schema          int    `json:"schema"`
	SessionID       string `json:"sessionId"`
	ArtifactPath    string `json:"artifactPath"`
	FormatVersion   int    `json:"formatVersion"`
	Size            int64  `json:"size"`
	ModTimeUnixNano int64  `json:"modTimeUnixNano"`
	// Next is the frame boundary the stream resumes from.
	Next int64 `json:"next"`
	// LastSeq is the last row read from the stream, so a gap is detectable.
	LastSeq int64 `json:"lastSeq"`
	// CaughtUp reports that the last read reached a clean end of file.
	CaughtUp  bool                      `json:"caughtUp,omitempty"`
	Consumers map[string]consumerRecord `json:"consumers,omitempty"`
}

// consumerRecord is one consumer's position and state.
type consumerRecord struct {
	Version int    `json:"version"`
	Seq     int64  `json:"seq"`
	State   []byte `json:"state,omitempty"`
}

// newRecord starts a record for an artifact with every consumer at the stream
// origin.
func newRecord(id string, ref sessions.LogRef, consumers []Consumer) *record {
	rec := &record{
		Schema:          recordSchema,
		SessionID:       id,
		ArtifactPath:    ref.Path,
		FormatVersion:   ref.FormatVersion,
		Size:            ref.Size,
		ModTimeUnixNano: ref.ModTimeUnixNano,
		Next:            0,
		LastSeq:         -1,
	}
	rec.ensure(consumers)
	return rec
}

// valid reports whether the cached record still describes this artifact. A new
// file at the same path, a new format generation, a truncated log, or a rewrite
// that did not change the size all invalidate it.
func (r *record) valid(id string, ref sessions.LogRef) bool {
	if r == nil || r.Schema != recordSchema || r.SessionID != id {
		return false
	}
	if r.ArtifactPath != ref.Path || r.FormatVersion != ref.FormatVersion {
		return false
	}
	if ref.Size < r.Size || r.Next > ref.Size {
		return false
	}
	if ref.Size == r.Size && r.Size > 0 && ref.ModTimeUnixNano != r.ModTimeUnixNano {
		return false
	}
	return true
}

// ensure reconciles the record with the registered consumers. A consumer with
// no cached state, or a different Version, starts from the stream origin; when
// that leaves a consumer behind the stream cursor, the stream is rewound so the
// lagging consumer can catch up while the others skip. It reports whether the
// stream was rewound.
func (r *record) ensure(consumers []Consumer) bool {
	if r.Consumers == nil {
		r.Consumers = make(map[string]consumerRecord, len(consumers))
	}
	registered := make(map[string]bool, len(consumers))
	for _, consumer := range consumers {
		id := consumer.ID()
		registered[id] = true
		entry, ok := r.Consumers[id]
		if !ok || entry.Version != consumer.Version() {
			r.Consumers[id] = consumerRecord{Version: consumer.Version(), Seq: -1}
		}
	}
	for id := range r.Consumers {
		if !registered[id] {
			delete(r.Consumers, id)
		}
	}
	if r.LastSeq >= 0 && r.slowest() < r.LastSeq {
		r.Next = 0
		r.LastSeq = -1
		return true
	}
	return false
}

// covers reports whether the record holds state for every registered consumer,
// at the version that consumer folds with. A record written by an older build
// can lack one: reading its absent state as zero would report a real-looking
// answer, so callers treat an uncovered record as no cache at all.
func (r *record) covers(consumers []Consumer) bool {
	if r == nil {
		return false
	}
	for _, consumer := range consumers {
		entry, ok := r.Consumers[consumer.ID()]
		if !ok || entry.Version != consumer.Version() {
			return false
		}
	}
	return true
}

// slowest is the position of the least advanced consumer.
func (r *record) slowest() int64 {
	if len(r.Consumers) == 0 {
		return r.LastSeq
	}
	slowest := int64(-1)
	first := true
	for _, entry := range r.Consumers {
		if first || entry.Seq < slowest {
			slowest = entry.Seq
			first = false
		}
	}
	return slowest
}

// snapshot builds the caller-facing view.
func (r *record) snapshot(readBytes, folded, skipped int64) Snapshot {
	values := make(map[string][]byte, len(r.Consumers))
	for id, entry := range r.Consumers {
		if len(entry.State) == 0 {
			continue
		}
		values[id] = entry.State
	}
	return Snapshot{
		SessionID:     r.SessionID,
		ArtifactPath:  r.ArtifactPath,
		FormatVersion: r.FormatVersion,
		Size:          r.Size,
		LastSeq:       r.LastSeq,
		CaughtUp:      r.CaughtUp,
		ReadBytes:     readBytes,
		Folded:        folded,
		Skipped:       skipped,
		Values:        values,
	}
}

// recordStore persists records beside the session cache. It never writes into
// the session directory itself, so removing the cache directory is always safe.
type recordStore struct {
	dir string
}

func (s recordStore) path(id string) string {
	return s.dir + string(os.PathSeparator) + encodeID(id) + ".json"
}

// load returns a validated record or nil when there is no usable cache.
func (s recordStore) load(id string) (*record, error) {
	data, err := os.ReadFile(s.path(id))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var rec record
	if err := jsonUnmarshal(data, &rec); err != nil {
		// A cache that cannot be parsed is simply a cache miss.
		return nil, nil
	}
	if rec.Schema != recordSchema || rec.SessionID != id {
		return nil, nil
	}
	return &rec, nil
}

// save writes atomically: a reader never sees a half-written record.
func (s recordStore) save(rec *record) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	data, err := jsonMarshal(rec)
	if err != nil {
		return err
	}
	path := s.path(rec.SessionID)
	tmp, err := os.CreateTemp(s.dir, ".record-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return err
	}
	if err := os.Rename(name, path); err != nil {
		os.Remove(name)
		return err
	}
	return nil
}

func (s recordStore) remove(id string) error {
	err := os.Remove(s.path(id))
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	return err
}

// encodeID keeps cache file names predictable without splitting on separators.
func encodeID(id string) string {
	if id == "" {
		return "session"
	}
	safe := make([]byte, 0, len(id))
	for i := 0; i < len(id); i++ {
		c := id[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-', c == '_', c == '.':
			safe = append(safe, c)
		default:
			safe = append(safe, '_')
		}
	}
	// Keep names well under the usual 255-byte limit.
	if len(safe) > 120 {
		safe = append(safe[:120], []byte(base64.RawURLEncoding.EncodeToString([]byte(id)))...)
	}
	return string(safe)
}

func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
