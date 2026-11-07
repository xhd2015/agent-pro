package sessions

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
	"time"
)

const defaultRecentLimit = 40

// ListRecent returns newest-first header summaries for at most limit
// sessions. It stats every artifact, then reads only those headers.
// It never calls readArtifact. limit <= 0 means 40.
func (s *Store) ListRecent(limit int) ([]Summary, []string, error) {
	if limit <= 0 {
		limit = defaultRecentLimit
	}
	artifacts, warnings, err := s.artifacts()
	if err != nil {
		return nil, nil, err
	}
	type scored struct {
		a    artifact
		mod  time.Time
		size int64
	}
	rows := make([]scored, 0, len(artifacts))
	for _, a := range artifacts {
		info, statErr := os.Lstat(a.path)
		if statErr != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", a.path, statErr))
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}
		rows = append(rows, scored{a: a, mod: info.ModTime(), size: info.Size()})
	}
	sort.Slice(rows, func(i, j int) bool {
		if !rows[i].mod.Equal(rows[j].mod) {
			return rows[i].mod.After(rows[j].mod)
		}
		return rows[i].a.path < rows[j].a.path
	})
	if len(rows) > limit {
		rows = rows[:limit]
	}
	type decoded struct {
		sum  Summary
		err  error
		path string
		mod  time.Time
		size int64
	}
	got := make([]decoded, len(rows))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for i, row := range rows {
		wg.Add(1)
		go func(i int, row scored) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			sum, err := cachedRecentSummary(row.a, row.size, row.mod)
			got[i] = decoded{sum: sum, err: err, path: row.a.path, mod: row.mod, size: row.size}
		}(i, row)
	}
	wg.Wait()
	out := make([]Summary, 0, len(rows))
	ids := map[string]string{}
	for _, row := range got {
		if row.err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", row.path, row.err))
			continue
		}
		sum := row.sum
		if prev, ok := ids[sum.ID]; ok {
			return nil, warnings, fmt.Errorf("duplicate session id %q in %s and %s", sum.ID, prev, row.path)
		}
		ids[sum.ID] = row.path
		sum.UpdatedAt = row.mod.UnixMilli()
		sum.SizeBytes = row.size
		out = append(out, sum)
	}
	return out, warnings, nil
}

type recentHeaderCacheEntry struct {
	size    int64
	modUnix int64
	sum     Summary
	// stable means the sample covered the lines the header and title live in,
	// so later appends cannot change them and the entry survives growth.
	stable bool
}

// recentHeaderCache keeps header summaries keyed by path, size, and mtime.
// ListRecent re-stats every log so order stays fresh, and only re-reads a
// file whose bytes changed.
var recentHeaderCache = struct {
	sync.Mutex
	items map[string]recentHeaderCacheEntry
}{items: map[string]recentHeaderCacheEntry{}}

func cachedRecentSummary(a artifact, size int64, mod time.Time) (Summary, error) {
	modUnix := mod.UnixNano()
	recentHeaderCache.Lock()
	if ent, ok := recentHeaderCache.items[a.path]; ok {
		if (ent.size == size && ent.modUnix == modUnix) || (ent.stable && size > ent.size) {
			sum := ent.sum
			recentHeaderCache.Unlock()
			return sum, nil
		}
	}
	recentHeaderCache.Unlock()
	sum, stable, err := recentSummaryStable(a)
	if err != nil {
		return Summary{}, err
	}
	recentHeaderCache.Lock()
	if len(recentHeaderCache.items) > 4000 {
		recentHeaderCache.items = map[string]recentHeaderCacheEntry{}
	}
	recentHeaderCache.items[a.path] = recentHeaderCacheEntry{size: size, modUnix: modUnix, sum: sum, stable: stable}
	recentHeaderCache.Unlock()
	return sum, nil
}

const recentPrefixBytes = 256 * 1024

// recentStableLines is the header plus the lines a title can appear in.
// A sample holding every one of them cannot be changed by an append.
const recentStableLines = 257

var newlineBytes = []byte("\n")

func recentSummary(a artifact) (Summary, error) {
	sum, _, err := recentSummaryStable(a)
	return sum, err
}

// recentSummaryStable reports whether the sampled prefix already contains
// every line that can affect the summary.
func recentSummaryStable(a artifact) (Summary, bool, error) {
	if !supportedFormat(a.version) {
		return Summary{}, false, unsupportedFormatError(a.version)
	}
	f, err := openArtifact(a.path)
	if err != nil {
		return Summary{}, false, err
	}
	defer f.Close()
	if a.encoding != "zstd" {
		prefix := make([]byte, recentPrefixBytes)
		n, err := f.Read(prefix)
		if err != nil && err != io.EOF {
			return Summary{}, false, err
		}
		return stableSummary(prefix[:n], a)
	}
	// The title lives in the first lines. Stop there instead of decoding
	// every frame in the 256KiB prefix. A torn frame after those lines is
	// the writer still appending; the header we already have is enough.
	dec, err := zstdStream(io.LimitReader(f, recentPrefixBytes))
	if err != nil {
		return Summary{}, false, err
	}
	defer releaseZstdStream(dec)
	var plain bytes.Buffer
	reader := bufio.NewReader(dec)
	for lines := 0; lines < recentStableLines; lines++ {
		line, readErr := reader.ReadBytes('\n')
		plain.Write(line)
		if readErr != nil {
			break
		}
	}
	if plain.Len() == 0 {
		return Summary{}, false, fmt.Errorf("incomplete first Zstandard frame")
	}
	return stableSummary(plain.Bytes(), a)
}

// stableSummary parses the sampled prefix and reports whether an append can
// still change it: only a sample holding every title line with a title found
// is final.
func stableSummary(sample []byte, a artifact) (Summary, bool, error) {
	sum, err := summaryFromPrefix(sample, a)
	if err != nil {
		return Summary{}, false, err
	}
	stable := sum.Title != "" && bytes.Count(sample, newlineBytes) >= recentStableLines
	return sum, stable, nil
}

func summaryFromPrefix(data []byte, a artifact) (Summary, error) {
	reader := bufio.NewReader(bytes.NewReader(data))
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return Summary{}, err
	}
	sum, err := parseSessionHeader(line, int(a.version), a.path, a.encoding)
	if err != nil {
		return Summary{}, err
	}
	for i := 0; i < 256; i++ {
		line, err = reader.ReadBytes('\n')
		if err != nil {
			break
		}
		if title := sessionTitleFromLine(line); title != "" {
			sum.Title = title
		}
	}
	return sum, nil
}

func sessionTitleFromLine(line []byte) string {
	fields, err := object(bytes.TrimSpace(line))
	if err != nil {
		return ""
	}
	typ, err := requiredString(fields, "type")
	if err != nil || typ != "session/title" {
		return ""
	}
	data, err := object(fields["data"])
	if err != nil {
		return ""
	}
	title, err := requiredString(data, "title")
	if err != nil {
		return ""
	}
	return title
}
