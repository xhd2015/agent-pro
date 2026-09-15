package usage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SnapshotSuffix is the file name suffix for stored snapshots.
const SnapshotSuffix = "-snapshot.jsonl"

// Store writes and reads usage snapshots under a usages root, laid out as
// <root>/<provider>/<YYYY-MM-DD>/<HH-MM-SS>-snapshot.jsonl. Times in paths are
// local, so the tree is browsable with ls; record timestamps inside are UTC.
type Store struct {
	Root string
}

// NewStore returns a store rooted at root (typically $AGENT_PRO_HOME/usages).
func NewStore(root string) *Store {
	return &Store{Root: root}
}

// Append writes one record and returns the file path it went to. A second
// snapshot within the same second appends another line to the same file, so
// runs are never overwritten.
func (s *Store) Append(rec Record) (string, error) {
	if strings.TrimSpace(s.Root) == "" {
		return "", fmt.Errorf("usage store: empty root")
	}
	line, err := rec.MarshalJSONLine()
	if err != nil {
		return "", fmt.Errorf("encode snapshot: %w", err)
	}

	dir := s.DayDir(rec.Provider, rec.LocalTime())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create usage dir: %w", err)
	}
	path := filepath.Join(dir, rec.LocalTime().Format("15-04-05")+SnapshotSuffix)

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return "", fmt.Errorf("open snapshot: %w", err)
	}
	defer file.Close()
	if _, err := file.Write(append(line, '\n')); err != nil {
		return "", fmt.Errorf("append snapshot: %w", err)
	}
	return path, nil
}

// DayDir returns the directory holding one provider's snapshots for a day.
func (s *Store) DayDir(provider ProviderID, at time.Time) string {
	return filepath.Join(s.Root, string(provider), at.Format("2006-01-02"))
}

// StoredRecord is a record read back from the store with its provenance.
type StoredRecord struct {
	Record Record
	Path   string
}

// FileFailure is one snapshot file that could not be read, with the path so a
// caller can report or skip it.
type FileFailure struct {
	Path string
	Err  error
}

// ListOptions selects what ListFiltered reads.
type ListOptions struct {
	// Last limits the records returned, newest first; 0 reads all of them.
	Last int
	// Since drops records collected before it. Day directories that cannot hold
	// an in-window record are skipped without being read. Zero reads the whole
	// tree.
	Since time.Time
}

// List returns up to last stored records, newest first, across every provider.
// last <= 0 means all of them. Files are visited newest first and reading stops
// once enough records are collected, always on an instant boundary. A file that
// cannot be read fails the whole call.
func (s *Store) List(last int) ([]StoredRecord, error) {
	out, failures, err := s.ListFiltered(ListOptions{Last: last})
	if err != nil {
		return nil, err
	}
	if len(failures) > 0 {
		return nil, failures[0].Err
	}
	return out, nil
}

// ListFiltered is List with a time bound, and it reports unreadable snapshot
// files instead of failing: a dashboard should chart the files it can read.
// Failures are in visit order, which is newest instant first.
func (s *Store) ListFiltered(opts ListOptions) ([]StoredRecord, []FileFailure, error) {
	paths, err := s.snapshotPaths(opts.Since)
	if err != nil {
		return nil, nil, err
	}

	var out []StoredRecord
	var failures []FileFailure
	for _, group := range snapshotGroups(paths) {
		for _, path := range group {
			records, err := readSnapshotFile(path)
			if err != nil {
				failures = append(failures, FileFailure{Path: path, Err: err})
				continue
			}
			// Within one file the last line is the newest snapshot.
			for i := len(records) - 1; i >= 0; i-- {
				if !opts.Since.IsZero() && records[i].TS.Before(opts.Since) {
					continue
				}
				out = append(out, StoredRecord{Record: records[i], Path: path})
			}
		}
		// Reading stops between instants: a snapshot one second older must never
		// displace a newer one just because its provider sorts earlier.
		if opts.Last > 0 && len(out) >= opts.Last {
			break
		}
	}

	// A file name carries whole seconds, so records collected within the same
	// second need the timestamp sort (then provider) to order stably.
	sort.SliceStable(out, func(i, j int) bool {
		if !out[i].Record.TS.Equal(out[j].Record.TS) {
			return out[i].Record.TS.After(out[j].Record.TS)
		}
		return out[i].Record.Provider < out[j].Record.Provider
	})
	if opts.Last > 0 && len(out) > opts.Last {
		out = out[:opts.Last]
	}
	return out, failures, nil
}

// snapshotGroups groups snapshot files by the instant they hold, newest instant
// first. One group is one second across every provider, so a caller can stop on
// a group boundary and still have the newest records. Paths inside a group are
// ordered by provider name, matching the provider tiebreak in List.
func snapshotGroups(paths []string) [][]string {
	byInstant := map[string][]string{}
	var instants []string
	for _, path := range paths {
		instant := filepath.Base(filepath.Dir(path)) + "/" + filepath.Base(path)
		if _, ok := byInstant[instant]; !ok {
			instants = append(instants, instant)
		}
		byInstant[instant] = append(byInstant[instant], path)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(instants)))

	groups := make([][]string, 0, len(instants))
	for _, instant := range instants {
		group := byInstant[instant]
		sort.Strings(group)
		groups = append(groups, group)
	}
	return groups
}

// snapshotPaths lists stored snapshot files, skipping the day directories that
// cannot hold a record collected at or after since. Day directory names are
// local dates while record timestamps are UTC, so the cutoff keeps one extra day.
func (s *Store) snapshotPaths(since time.Time) ([]string, error) {
	if strings.TrimSpace(s.Root) == "" {
		return nil, nil
	}
	cutoff := ""
	if !since.IsZero() {
		cutoff = since.In(time.Local).AddDate(0, 0, -1).Format("2006-01-02")
	}
	var paths []string
	providers, err := os.ReadDir(s.Root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	for _, provider := range providers {
		if !provider.IsDir() {
			continue
		}
		days, err := os.ReadDir(filepath.Join(s.Root, provider.Name()))
		if err != nil {
			continue
		}
		for _, day := range days {
			if !day.IsDir() || beforeCutoff(day.Name(), cutoff) {
				continue
			}
			files, err := os.ReadDir(filepath.Join(s.Root, provider.Name(), day.Name()))
			if err != nil {
				continue
			}
			for _, file := range files {
				if file.IsDir() || !strings.HasSuffix(file.Name(), SnapshotSuffix) {
					continue
				}
				paths = append(paths, filepath.Join(s.Root, provider.Name(), day.Name(), file.Name()))
			}
		}
	}
	return paths, nil
}

// beforeCutoff reports whether a day directory name is older than the cutoff
// date. Names that are not dates are always visited.
func beforeCutoff(name, cutoff string) bool {
	if cutoff == "" || name >= cutoff {
		return false
	}
	if _, err := time.Parse("2006-01-02", name); err != nil {
		return false
	}
	return true
}

func readSnapshotFile(path string) ([]Record, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var records []Record
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var rec Record
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		records = append(records, rec)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return records, nil
}

// LocalTime is the record timestamp in local time, used for file names.
func (r Record) LocalTime() time.Time {
	return r.TS.Local()
}
