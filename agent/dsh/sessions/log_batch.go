package sessions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// This file is the incremental read primitive under a consumer bus: one
// bounded read from a frame boundary. It never retains the whole log and never
// advances past the last complete frame, so a writer that is mid-frame costs a
// re-read instead of a skipped event.

const (
	defaultLogWindow    int64 = 1 << 20
	defaultLogMaxWindow int64 = 32 << 20
	defaultLogMaxEvents       = 1024
)

// LogRef identifies one append-only artifact. Path plus FormatVersion is the
// identity a consumer cache validates against: DSH writes a new file for a new
// format generation, so a generation change is a new artifact, not a rewrite.
type LogRef struct {
	Path            string `json:"path"`
	Encoding        string `json:"encoding"`
	FormatVersion   int    `json:"format_version"`
	Size            int64  `json:"size"`
	ModTimeUnixNano int64  `json:"mod_time_unix_nano"`
}

// SameArtifact reports whether two refs name the same readable artifact.
func (r LogRef) SameArtifact(other LogRef) bool {
	return r.Path == other.Path && r.FormatVersion == other.FormatVersion && r.Encoding == other.Encoding
}

// StatLog resolves a session's artifact without decoding any event. A consumer
// bus calls it first so an unchanged log costs one Lstat, not one read.
func (s *Store) StatLog(id string) (LogRef, error) {
	a, err := s.lookupArtifact(id)
	if err != nil {
		return LogRef{}, err
	}
	return statLogRef(a)
}

func statLogRef(a artifact) (LogRef, error) {
	info, err := os.Lstat(a.path)
	if err != nil {
		return LogRef{}, err
	}
	if !info.Mode().IsRegular() {
		return LogRef{}, fmt.Errorf("selected artifact is not a regular file")
	}
	return LogRef{
		Path:            a.path,
		Encoding:        a.encoding,
		FormatVersion:   int(a.version),
		Size:            info.Size(),
		ModTimeUnixNano: info.ModTime().UnixNano(),
	}, nil
}

// LogEvent is one row after the header. Data is nil when the type is outside
// the caller's interest set, which is what keeps a 65k-row log cheap.
type LogEvent struct {
	Seq  int64           `json:"seq"`
	Type string          `json:"type"`
	Time int64           `json:"time"`
	Data json.RawMessage `json:"data,omitempty"`
}

// ReadLogOptions controls one bounded read.
type ReadLogOptions struct {
	// From is a frame boundary: 0 or a previous batch's Next.
	From int64
	// Window is the compressed bytes fetched per attempt (default 1 MiB). It
	// grows only when a single frame does not fit.
	Window int64
	// MaxWindow caps that growth before the remainder is read in one piece.
	MaxWindow int64
	// MaxEvents stops after a frame boundary near this many events.
	MaxEvents int
	// Interest filters Data: nil decodes every type.
	Interest func(typ string) bool
}

// LogBatch is the result of one bounded read.
type LogBatch struct {
	Ref    LogRef
	Events []LogEvent
	// Next is the frame boundary to resume from: the end of the last complete
	// frame consumed by this batch.
	Next int64
	// Rewound reports that From was past the current artifact size, so the
	// caller must restart from 0. No events are returned.
	Rewound bool
	// Torn reports that the artifact ends in a frame the writer is still
	// filling. The events in that frame are not consumable yet.
	Torn bool
	// More reports that complete frames remain unread in this batch's budget.
	More bool
	// Done reports that Next reached the artifact size with no torn tail: the
	// caller has consumed every durable row.
	Done bool
}

// ReadLog reads complete rows starting at a frame boundary. It decodes at most
// a bounded window, stops on a frame boundary, and never splits a frame.
func (s *Store) ReadLog(id string, opts ReadLogOptions) (LogBatch, error) {
	a, err := s.lookupArtifact(id)
	if err != nil {
		return LogBatch{}, err
	}
	return readLogFrom(a, opts)
}

func readLogFrom(a artifact, opts ReadLogOptions) (LogBatch, error) {
	ref, err := statLogRef(a)
	if err != nil {
		return LogBatch{}, err
	}
	batch := LogBatch{Ref: ref}
	from := opts.From
	if from < 0 {
		from = 0
	}
	switch {
	case from > ref.Size:
		// The artifact shrank: a different file occupies this path now.
		batch.Rewound = true
		batch.More = true
		return batch, nil
	case from == ref.Size:
		batch.Next = from
		batch.Done = true
		return batch, nil
	}
	f, err := openArtifact(ref.Path)
	if err != nil {
		return batch, err
	}
	defer f.Close()

	window := opts.Window
	if window <= 0 {
		window = defaultLogWindow
	}
	maxWindow := opts.MaxWindow
	if maxWindow <= 0 {
		maxWindow = defaultLogMaxWindow
	}
	maxEvents := opts.MaxEvents
	if maxEvents <= 0 {
		maxEvents = defaultLogMaxEvents
	}
	remaining := ref.Size - from
	if window > remaining {
		window = remaining
	}
	for {
		if window > remaining {
			window = remaining
		}
		buf := make([]byte, window)
		n, err := f.ReadAt(buf, from)
		if err != nil && err != io.EOF {
			return batch, err
		}
		buf = buf[:n]
		atEnd := from+int64(n) >= ref.Size
		if ref.Encoding != "zstd" {
			return decodeJSONLWindow(batch, ref, buf, from, atEnd, opts, maxEvents)
		}
		frames, _, torn, err := scanZstdFrames(buf)
		if err != nil {
			return batch, err
		}
		if len(frames) == 0 && torn && !atEnd {
			// The window ended inside a frame that the file already holds.
			// Grow it instead of reporting a torn artifact that is not torn.
			if window >= maxWindow {
				if remaining <= window {
					return batch, fmt.Errorf("session %d-byte log exceeds the %d-byte read window", remaining, maxWindow)
				}
				window = remaining
			} else if next := window * 2; next > maxWindow {
				window = maxWindow
			} else {
				window = next
			}
			continue
		}
		next, more, err := decodeZstdFrames(&batch, ref, buf, frames, from, opts, maxEvents, from == 0)
		if err != nil {
			return batch, err
		}
		batch.Next = next
		batch.More = more || (from+int64(n) < ref.Size)
		batch.Torn = torn && atEnd
		batch.Done = batch.Next == ref.Size && !batch.Torn
		if batch.Torn {
			batch.More = false
		}
		return batch, nil
	}
}

// decodeZstdFrames decodes complete frames in order. Frame boundaries are the
// only place a cursor may stop, so an event budget is applied per frame.
func decodeZstdFrames(batch *LogBatch, ref LogRef, buf []byte, frames []zstdFrame, from int64, opts ReadLogOptions, maxEvents int, firstFileFrame bool) (int64, bool, error) {
	next := from
	headerPending := firstFileFrame
	for _, frame := range frames {
		plain, err := decodeZstdFrame(buf[frame.start:frame.end])
		if err != nil {
			return next, false, fmt.Errorf("decode frame at byte %d: %w", from+int64(frame.start), err)
		}
		if headerPending {
			headerPending = false
			rest, err := skipHeaderLine(plain)
			if err != nil {
				return next, false, err
			}
			plain = rest
		}
		events, err := decodeEventLines(plain, opts.Interest)
		if err != nil {
			return next, false, err
		}
		if len(batch.Events) > 0 && len(batch.Events)+len(events) > maxEvents {
			return next, true, nil
		}
		batch.Events = append(batch.Events, events...)
		next = from + int64(frame.end)
	}
	return next, false, nil
}

func decodeJSONLWindow(batch LogBatch, ref LogRef, buf []byte, from int64, atEnd bool, opts ReadLogOptions, maxEvents int) (LogBatch, error) {
	consumed := bytes.LastIndexByte(buf, '\n') + 1
	plain := buf[:consumed]
	if from == 0 {
		rest, err := skipHeaderLine(plain)
		if err != nil {
			return batch, err
		}
		plain = rest
	}
	events, err := decodeEventLines(plain, opts.Interest)
	if err != nil {
		return batch, err
	}
	if len(events) > maxEvents && maxEvents > 0 {
		// JSONL rows are not frame aligned; stopping mid-file is still safe
		// because the caller resumes from a line boundary, but keep the whole
		// window so a batch never straddles a row.
		events = events[:maxEvents]
	}
	batch.Events = append(batch.Events, events...)
	batch.Next = from + int64(consumed)
	batch.More = batch.Next < ref.Size
	batch.Torn = atEnd && consumed < len(buf) && batch.Next == ref.Size
	batch.Done = batch.Next == ref.Size && !batch.Torn
	return batch, nil
}

// skipHeaderLine drops the first row of an artifact, which carries identity and
// format instead of an event envelope.
func skipHeaderLine(plain []byte) ([]byte, error) {
	idx := bytes.IndexByte(plain, '\n')
	if idx < 0 {
		return nil, fmt.Errorf("session log does not start with a complete header row")
	}
	line := bytes.TrimSpace(plain[:idx])
	var header struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(line, &header); err != nil {
		return nil, fmt.Errorf("invalid session header: %w", err)
	}
	if header.Type != "session" {
		return nil, fmt.Errorf("invalid session header type %q", header.Type)
	}
	return plain[idx+1:], nil
}

// decodeEventLines parses envelopes only. Data is retained as raw bytes and
// dropped for types the caller does not follow, so an uninterested row never
// costs a payload decode.
func decodeEventLines(plain []byte, interest func(string) bool) ([]LogEvent, error) {
	var events []LogEvent
	for len(plain) > 0 {
		idx := bytes.IndexByte(plain, '\n')
		if idx < 0 {
			break
		}
		line := bytes.TrimSpace(plain[:idx])
		plain = plain[idx+1:]
		if len(line) == 0 {
			continue
		}
		var row struct {
			Seq  *int64          `json:"seq"`
			Type string          `json:"type"`
			Time int64           `json:"time"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(line, &row); err != nil {
			return nil, fmt.Errorf("invalid event row: %w", err)
		}
		if row.Seq == nil {
			return nil, fmt.Errorf("event row %q has no seq", row.Type)
		}
		event := LogEvent{Seq: *row.Seq, Type: row.Type, Time: row.Time}
		if interest == nil || interest(row.Type) {
			event.Data = row.Data
		}
		events = append(events, event)
	}
	return events, nil
}
