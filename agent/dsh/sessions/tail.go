package sessions

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/klauspost/compress/zstd"
)

const tailKeep = 200

// coldTailMin is the compressed size above which the first SyncTail reads the
// header and a suffix instead of JSON-parsing the whole log. Smaller logs,
// including the session fixtures, stay on the full read.
const coldTailMin = 256 * 1024

// coldTailWindow is the first suffix. It doubles until the suffix holds
// coldTailRows transcript rows and a classify anchor, or it covers the file.
const coldTailWindow = 256 * 1024

// coldTailRows is the suffix size that satisfies a live transcript poll.
// The orchestrator asks for 24 rows; tailKeep is only the in-memory cap.
const coldTailRows = 24

// Tail is an incremental session cursor. It never retains the full event log.
type Tail struct {
	Path      string
	Size      int64
	Offset    int64
	Summary   Summary
	Rows      []TranscriptRow
	fold      classifyState
	nextSeq   int64
	header    bool
	resyncSeq bool
	windowed  bool
	sawAnchor bool
}

// SyncTail reads only bytes after the previous offset (new jsonl lines or
// new zstd frames). prev may be nil. It does not call readArtifact.
func (s *Store) SyncTail(id string, prev *Tail) (*Tail, error) {
	a, err := s.lookupArtifact(id)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(a.path)
	if err != nil {
		return nil, err
	}
	if prev != nil && prev.Path == a.path && prev.Size == info.Size() {
		return prev, nil
	}
	t := prev
	if t == nil || t.Path != a.path || info.Size() < t.Size {
		t = &Tail{Path: a.path, Summary: Summary{Path: a.path, Encoding: a.encoding, FormatVersion: int(a.version), Outcome: "empty"}}
	}
	if err := t.advance(a, info.Size()); err != nil {
		return nil, err
	}
	t.Size = info.Size()
	t.Summary.SizeBytes = info.Size()
	return t, nil
}

func (t *Tail) Activity(live, liveKnown bool) Activity {
	return t.fold.finish(live, liveKnown)
}

func (t *Tail) advance(a artifact, size int64) error {
	if size <= t.Offset {
		return nil
	}
	if t.Offset == 0 && !t.header && a.encoding == "zstd" && size > coldTailMin {
		return t.advanceColdZstd(a, size)
	}
	return t.advanceBytes(a, size)
}

// advanceColdZstd reads the session header from the front of a large
// concatenated-zstd log and folds only a suffix. The cursor stays on a frame
// boundary, so a later append still parses just the new frames.
func (t *Tail) advanceColdZstd(a artifact, size int64) error {
	base, err := recentSummary(a)
	if err != nil {
		return err
	}
	window := int64(coldTailWindow)
	for {
		if window > size {
			window = size
		}
		start, ok, err := suffixFrameStart(a.path, size, window)
		if err != nil {
			return err
		}
		if ok && start > 0 {
			probe := &Tail{
				Path:      a.path,
				Offset:    start,
				Summary:   base,
				header:    true,
				resyncSeq: true,
				windowed:  true,
			}
			if err := probe.advanceBytes(a, size); err != nil {
				return err
			}
			if probe.enoughTail() || window == size {
				*t = *probe
				return nil
			}
		}
		if window == size {
			break
		}
		window *= 2
	}
	t.Path = a.path
	t.Offset = 0
	t.Summary = Summary{Path: a.path, Encoding: a.encoding, FormatVersion: int(a.version), Outcome: "empty"}
	t.Rows = nil
	t.fold = classifyState{}
	t.nextSeq = 0
	t.header = false
	t.resyncSeq = false
	t.windowed = false
	t.sawAnchor = false
	return t.advanceBytes(a, size)
}

func (t *Tail) enoughTail() bool {
	return t.sawAnchor && len(t.Rows) >= coldTailRows
}

func suffixFrameStart(path string, size, window int64) (int64, bool, error) {
	if window > size {
		window = size
	}
	from := size - window
	f, err := openArtifact(path)
	if err != nil {
		return 0, false, err
	}
	defer f.Close()
	buf := make([]byte, window)
	n, err := f.ReadAt(buf, from)
	if err != nil && err != io.EOF {
		return 0, false, err
	}
	rel := findFrameStart(buf[:n])
	if rel < 0 {
		return 0, false, nil
	}
	return from + int64(rel), true, nil
}

func (t *Tail) advanceBytes(a artifact, size int64) error {
	if size <= t.Offset {
		return nil
	}
	f, err := openArtifact(a.path)
	if err != nil {
		return err
	}
	defer f.Close()
	n := size - t.Offset
	buf := make([]byte, n)
	got, err := f.ReadAt(buf, t.Offset)
	if err != nil && err != io.EOF {
		return err
	}
	buf = buf[:got]
	if a.encoding == "zstd" {
		return t.advanceZstd(buf)
	}
	return t.advanceJSONL(buf)
}

func (t *Tail) advanceZstd(suffix []byte) error {
	frames, tornStart, torn, err := scanZstdFrames(suffix)
	if err != nil {
		return err
	}
	for _, frame := range frames {
		plain, err := decodeZstdFrame(suffix[frame.start:frame.end])
		if err != nil {
			return err
		}
		if _, err := t.consumeJSONL(plain, true); err != nil {
			return err
		}
	}
	if torn {
		t.Offset += int64(tornStart)
		return nil
	}
	t.Offset += int64(len(suffix))
	return nil
}

func (t *Tail) advanceJSONL(suffix []byte) error {
	consumed, err := t.consumeJSONL(suffix, false)
	if err != nil {
		return err
	}
	t.Offset += int64(consumed)
	return nil
}

func decodeZstdFrame(frame []byte) ([]byte, error) {
	dec, err := zstd.NewReader(bytes.NewReader(frame), zstd.WithDecoderConcurrency(1))
	if err != nil {
		return nil, err
	}
	defer dec.Close()
	return io.ReadAll(dec)
}

func (t *Tail) consumeJSONL(data []byte, complete bool) (int, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	consumed := 0
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			if len(line) == 0 {
				return consumed, nil
			}
			if !complete {
				return consumed, nil
			}
			return consumed, fmt.Errorf("incomplete JSONL row")
		}
		if err != nil {
			return consumed, err
		}
		consumed += len(line)
		if err := t.consumeLine(bytes.TrimSpace(line)); err != nil {
			return consumed, err
		}
	}
}

func lineSeq(line []byte) (int64, error) {
	fields, err := object(line)
	if err != nil {
		return 0, err
	}
	return requiredInteger(fields, "seq", true)
}

func tailAnchor(typ string) bool {
	switch typ {
	case "turn/start", "turn/end", "assistant/attempt", "assistant/message", "tool/call", "command/run":
		return true
	default:
		return false
	}
}

func (t *Tail) consumeLine(line []byte) error {
	if len(line) == 0 {
		return nil
	}
	if !t.header {
		sum, err := parseSessionHeader(line, t.Summary.FormatVersion, t.Path, t.Summary.Encoding)
		if err != nil {
			return err
		}
		sum.SizeBytes = t.Summary.SizeBytes
		sum.Outcome = t.Summary.Outcome
		if sum.Outcome == "" {
			sum.Outcome = "empty"
		}
		t.Summary = sum
		t.header = true
		return nil
	}
	if t.resyncSeq {
		seq, err := lineSeq(line)
		if err != nil {
			return err
		}
		t.nextSeq = seq
		t.resyncSeq = false
	}
	event, err := parseEvent(line, t.nextSeq)
	if err != nil {
		return fmt.Errorf("event row %d: %w", t.nextSeq+1, err)
	}
	t.nextSeq++
	if err := project(&t.Summary, event); err != nil {
		return err
	}
	if event.Time > t.Summary.UpdatedAt {
		t.Summary.UpdatedAt = event.Time
	}
	t.fold.add(event)
	if tailAnchor(event.Type) {
		t.sawAnchor = true
	}
	if t.windowed && t.fold.turnOpen && (t.Summary.Outcome == "" || t.Summary.Outcome == "empty") {
		t.Summary.Outcome = "incomplete"
	}
	if row, ok := transcriptRow(event); ok {
		if len(t.Rows) >= tailKeep {
			copy(t.Rows, t.Rows[1:])
			t.Rows[len(t.Rows)-1] = row
		} else {
			t.Rows = append(t.Rows, row)
		}
	}
	return nil
}

func parseSessionHeader(line []byte, version int, path, encoding string) (Summary, error) {
	header, err := object(line)
	if err != nil {
		return Summary{}, fmt.Errorf("invalid session header: %w", err)
	}
	typ, err := requiredString(header, "type")
	if err != nil || typ != "session" {
		return Summary{}, fmt.Errorf("invalid session header type")
	}
	ver, err := requiredInteger(header, "version", true)
	if err != nil || ver != int64(version) {
		return Summary{}, fmt.Errorf("header version does not match generation v%d", version)
	}
	id, err := requiredString(header, "id")
	if err != nil {
		return Summary{}, err
	}
	created, err := requiredInteger(header, "createdAt", true)
	if err != nil {
		return Summary{}, err
	}
	raw, ok := header["isSeeded"]
	if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return Summary{}, fmt.Errorf("isSeeded must be a boolean")
	}
	var seeded bool
	if err := json.Unmarshal(raw, &seeded); err != nil {
		return Summary{}, err
	}
	depth, err := requiredInteger(header, "delegationDepth", true)
	if err != nil {
		return Summary{}, err
	}
	preset, err := optionalString(header, "agentPreset")
	if err != nil {
		return Summary{}, err
	}
	cwd, err := optionalString(header, "cwd")
	if err != nil {
		return Summary{}, err
	}
	parent, err := optionalString(header, "parentSession")
	if err != nil {
		return Summary{}, err
	}
	origin, err := optionalString(header, "origin")
	if err != nil {
		return Summary{}, err
	}
	return Summary{
		ID: id, CreatedAt: created, UpdatedAt: created, Cwd: cwd,
		ParentSession: parent, Origin: origin, IsSeeded: seeded,
		AgentPreset: preset, DelegationDepth: depth, FormatVersion: version,
		Path: path, Encoding: encoding, Outcome: "empty",
	}, nil
}
