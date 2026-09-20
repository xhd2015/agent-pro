package sessions

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/klauspost/compress/zstd"
)

const maxSafeInteger int64 = 9007199254740991

// ResolveRoot resolves the explicit session root or the selected DSH home's sessions directory.
func ResolveRoot(opts Options) (string, error) {
	path := opts.SessionsDir
	if path == "" {
		home := opts.Home
		if home == "" {
			home = os.Getenv("DSH_HOME")
			if strings.TrimSpace(home) == "" {
				home = "~/.dsh"
			}
		}
		path = filepath.Join(home, "sessions")
	}
	if path == "~" || strings.HasPrefix(path, "~/") || strings.HasPrefix(path, `~\`) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			path = home
		} else {
			path = filepath.Join(home, path[2:])
		}
	}
	return filepath.Abs(path)
}

// Store reads a project/session directory tree rooted at Root.
type Store struct{ Root string }

// NewStore resolves the root without creating directories or reading logs.
func NewStore(opts Options) (*Store, error) {
	root, err := ResolveRoot(opts)
	if err != nil {
		return nil, err
	}
	return &Store{Root: root}, nil
}

type artifact struct {
	path, encoding string
	version        uint64
}

var generationName = regexp.MustCompile(`^session(?:\.v([1-9][0-9]*))?\.jsonl(\.zstd)?$`)

// artifacts selects generations before parsing headers. Encoding conflicts invalidate the whole root.
func (s *Store) artifacts() ([]artifact, []string, error) {
	projects, err := os.ReadDir(s.Root)
	if os.IsNotExist(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	var out []artifact
	var warnings []string
	encoding := ""
	for _, project := range projects {
		if !project.IsDir() {
			continue
		}
		projectPath := filepath.Join(s.Root, project.Name())
		dirs, err := os.ReadDir(projectPath)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", projectPath, err))
			continue
		}
		for _, dir := range dirs {
			if !dir.IsDir() {
				continue
			}
			dirPath := filepath.Join(projectPath, dir.Name())
			entries, err := os.ReadDir(dirPath)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("%s: %v", dirPath, err))
				continue
			}
			var selected *artifact
			invalid := false
			for _, entry := range entries {
				m := generationName.FindStringSubmatch(entry.Name())
				if m == nil {
					continue
				}
				enc := "jsonl"
				if m[2] != "" {
					enc = "zstd"
				}
				if encoding != "" && encoding != enc {
					return nil, nil, fmt.Errorf("mixed JSONL and Zstandard encodings under %s", s.Root)
				}
				encoding = enc
				version := uint64(0)
				if m[1] != "" {
					version, err = strconv.ParseUint(m[1], 10, 64)
					if err != nil {
						invalid = true
						warnings = append(warnings, fmt.Sprintf("%s: unsupported generation number %s", dirPath, m[1]))
						continue
					}
				}
				a := artifact{filepath.Join(dirPath, entry.Name()), enc, version}
				if selected == nil || version > selected.version {
					selected = &a
				}
			}
			if selected != nil && !invalid {
				out = append(out, *selected)
			}
		}
	}
	return out, warnings, nil
}

// List returns newest-first summaries, isolating unreadable sessions as warnings. Duplicate IDs are fatal.
func (s *Store) List() ([]Summary, []string, error) {
	artifacts, warnings, err := s.artifacts()
	if err != nil {
		return nil, nil, err
	}
	out := []Summary{}
	ids := map[string]string{}
	for _, a := range artifacts {
		id, identityErr := artifactID(a)
		if identityErr == nil {
			if prev, ok := ids[id]; ok {
				return nil, warnings, fmt.Errorf("duplicate session id %q in %s and %s", id, prev, a.path)
			}
			ids[id] = a.path
		}
		inspection, err := readArtifact(a)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", a.path, err))
			continue
		}
		out = append(out, inspection.Summary)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt != out[j].UpdatedAt {
			return out[i].UpdatedAt > out[j].UpdatedAt
		}
		return out[i].ID < out[j].ID
	})
	return out, warnings, nil
}

// Inspect matches an exact header identity, never interpolating the supplied ID into a path.
func (s *Store) Inspect(id string) (*Inspection, error) {
	if id == "" {
		return nil, fmt.Errorf("session id is required")
	}
	artifacts, warnings, err := s.artifacts()
	if err != nil {
		return nil, err
	}
	var selected *artifact
	var failures []string
	for _, a := range artifacts {
		actual, err := artifactID(a)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", a.path, err))
			continue
		}
		if actual != id {
			continue
		}
		if selected != nil {
			return nil, fmt.Errorf("duplicate session id %q in %s and %s", id, selected.path, a.path)
		}
		copy := a
		selected = &copy
	}
	if selected != nil {
		result, err := readArtifact(*selected)
		if err != nil {
			return nil, fmt.Errorf("session %q (%s): %w", id, selected.path, err)
		}
		if result.ID != id {
			return nil, fmt.Errorf("session header identity changed while reading %s", selected.path)
		}
		return result, nil
	}
	failures = append(warnings, failures...)
	detail := ""
	if len(failures) > 0 {
		detail = "; unreadable entries: " + strings.Join(failures, "; ")
	}
	return nil, fmt.Errorf("DSH session %q was not found under %s%s", id, s.Root, detail)
}

// Session artifacts must be ordinary files; links and special files are not inspected.
func openArtifact(path string) (*os.File, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("selected artifact is not a regular file")
	}
	return os.Open(path)
}

func artifactID(a artifact) (string, error) {
	f, err := openArtifact(a.path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return "", err
	}
	if !stat.Mode().IsRegular() {
		return "", fmt.Errorf("selected artifact is not regular")
	}
	var input io.Reader = f
	if a.encoding == "zstd" {
		dec, err := zstd.NewReader(f, zstd.WithDecoderConcurrency(1))
		if err != nil {
			return "", err
		}
		defer dec.Close()
		input = dec
	}
	line, err := bufio.NewReader(input).ReadBytes('\n')
	if err != nil {
		return "", err
	}
	header, err := object(line)
	if err != nil {
		return "", err
	}
	typ, err := requiredString(header, "type")
	if err != nil || typ != "session" {
		return "", fmt.Errorf("invalid session header type")
	}
	return requiredString(header, "id")
}

func readArtifact(a artifact) (*Inspection, error) {
	if a.version != 3 {
		return nil, fmt.Errorf("unsupported session format v%d (reader supports v3)", a.version)
	}
	f, err := openArtifact(a.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !stat.Mode().IsRegular() {
		return nil, fmt.Errorf("selected artifact is not a regular file")
	}
	var input io.Reader = f
	if a.encoding == "zstd" {
		dec, err := zstd.NewReader(f, zstd.WithDecoderConcurrency(1))
		if err != nil {
			return nil, err
		}
		defer dec.Close()
		input = dec
	}
	reader := bufio.NewReader(input)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("read complete session header: %w", err)
	}
	header, err := object(line)
	if err != nil {
		return nil, fmt.Errorf("invalid session header: %w", err)
	}
	typ, err := requiredString(header, "type")
	if err != nil || typ != "session" {
		return nil, fmt.Errorf("invalid session header type")
	}
	version, err := requiredInteger(header, "version", true)
	if err != nil || version != int64(a.version) {
		return nil, fmt.Errorf("header version does not match generation v%d", a.version)
	}
	id, err := requiredString(header, "id")
	if err != nil {
		return nil, err
	}
	created, err := requiredInteger(header, "createdAt", true)
	if err != nil {
		return nil, err
	}
	var seeded bool
	raw, ok := header["isSeeded"]
	if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, fmt.Errorf("isSeeded must be a boolean")
	}
	if err := json.Unmarshal(raw, &seeded); err != nil {
		return nil, err
	}
	depth, err := requiredInteger(header, "delegationDepth", true)
	if err != nil {
		return nil, err
	}
	preset, err := optionalString(header, "agentPreset")
	if err != nil {
		return nil, err
	}
	cwd, err := optionalString(header, "cwd")
	if err != nil {
		return nil, err
	}
	parent, err := optionalString(header, "parentSession")
	if err != nil {
		return nil, err
	}
	origin, err := optionalString(header, "origin")
	if err != nil {
		return nil, err
	}
	result := &Inspection{Summary: Summary{ID: id, CreatedAt: created, UpdatedAt: created, Cwd: cwd, ParentSession: parent, Origin: origin, IsSeeded: seeded, AgentPreset: preset, DelegationDepth: depth, FormatVersion: 3, Path: a.path, Encoding: a.encoding, SizeBytes: stat.Size(), Outcome: "empty"}, Header: bytes.Clone(bytes.TrimSpace(line)), Events: []Event{}}
	for {
		line, err = reader.ReadBytes('\n')
		if err == io.EOF {
			if len(line) > 0 && a.encoding == "zstd" {
				return nil, fmt.Errorf("complete compressed stream ends in an incomplete JSONL row")
			}
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read event: %w", err)
		}
		event, err := parseEvent(line, int64(len(result.Events)))
		if err != nil {
			return nil, fmt.Errorf("event row %d: %w", len(result.Events)+1, err)
		}
		if err := project(&result.Summary, event); err != nil {
			return nil, fmt.Errorf("event %d %s: %w", event.Seq, event.Type, err)
		}
		result.Events = append(result.Events, event)
		if event.Time > result.UpdatedAt {
			result.UpdatedAt = event.Time
		}
	}
	return result, nil
}

func object(raw []byte) (map[string]json.RawMessage, error) {
	var value map[string]json.RawMessage
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	if value == nil {
		return nil, fmt.Errorf("expected JSON object")
	}
	return value, nil
}
func requiredString(fields map[string]json.RawMessage, key string) (string, error) {
	var value string
	raw, ok := fields[key]
	if !ok || json.Unmarshal(raw, &value) != nil || value == "" {
		return "", fmt.Errorf("%s must be a non-empty string", key)
	}
	return value, nil
}
func optionalString(fields map[string]json.RawMessage, key string) (string, error) {
	if _, ok := fields[key]; !ok {
		return "", nil
	}
	return requiredString(fields, key)
}
func requiredInteger(fields map[string]json.RawMessage, key string, nonnegative bool) (int64, error) {
	var value int64
	raw, ok := fields[key]
	if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &value) != nil || value > maxSafeInteger || value < -maxSafeInteger || (nonnegative && value < 0) {
		return 0, fmt.Errorf("%s must be a safe integer", key)
	}
	return value, nil
}
func parseEvent(raw []byte, expected int64) (Event, error) {
	fields, err := object(raw)
	if err != nil {
		return Event{}, err
	}
	typ, err := requiredString(fields, "type")
	if err != nil {
		return Event{}, err
	}
	seq, err := requiredInteger(fields, "seq", true)
	if err != nil {
		return Event{}, err
	}
	if seq != expected {
		return Event{}, fmt.Errorf("expected seq %d, got %d", expected, seq)
	}
	timestamp, err := requiredInteger(fields, "time", false)
	if err != nil {
		return Event{}, err
	}
	data, ok := fields["data"]
	if !ok {
		return Event{}, fmt.Errorf("missing data")
	}
	return Event{Type: typ, Seq: seq, Time: timestamp, Data: bytes.Clone(data), Raw: bytes.Clone(bytes.TrimSpace(raw))}, nil
}
func project(summary *Summary, event Event) error {
	switch event.Type {
	case "turn/start", "turn/end", "request/header":
		data, err := object(event.Data)
		if err != nil {
			return err
		}
		switch event.Type {
		case "turn/start":
			if _, err := requiredInteger(data, "turn", true); err != nil {
				return err
			}
			summary.Outcome = "incomplete"
		case "turn/end":
			if _, err := requiredInteger(data, "turn", true); err != nil {
				return err
			}
			reason, err := object(data["reason"])
			if err != nil {
				return err
			}
			kind, err := requiredString(reason, "kind")
			if err != nil {
				return err
			}
			summary.Outcome = kind
		case "request/header":
			header, err := object(data["header"])
			if err != nil {
				return err
			}
			config, err := object(header["config"])
			if err != nil {
				return err
			}
			provider, err := requiredString(config, "provider")
			if err != nil {
				return err
			}
			model, err := requiredString(config, "model")
			if err != nil {
				return err
			}
			summary.Provider = provider
			summary.Model = model
		}
	}
	return nil
}
