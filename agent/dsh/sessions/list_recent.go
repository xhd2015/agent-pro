package sessions

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/klauspost/compress/zstd"
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
	out := make([]Summary, 0, len(rows))
	ids := map[string]string{}
	for _, row := range rows {
		sum, headErr := recentSummary(row.a)
		if headErr != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", row.a.path, headErr))
			continue
		}
		if prev, ok := ids[sum.ID]; ok {
			return nil, warnings, fmt.Errorf("duplicate session id %q in %s and %s", sum.ID, prev, row.a.path)
		}
		ids[sum.ID] = row.a.path
		sum.UpdatedAt = row.mod.UnixMilli()
		sum.SizeBytes = row.size
		out = append(out, sum)
	}
	return out, warnings, nil
}

const recentPrefixBytes = 256 * 1024

func recentSummary(a artifact) (Summary, error) {
	if !supportedFormat(a.version) {
		return Summary{}, unsupportedFormatError(a.version)
	}
	f, err := openArtifact(a.path)
	if err != nil {
		return Summary{}, err
	}
	defer f.Close()
	prefix := make([]byte, recentPrefixBytes)
	n, err := f.Read(prefix)
	if err != nil && err != io.EOF {
		return Summary{}, err
	}
	prefix = prefix[:n]
	plain := prefix
	if a.encoding == "zstd" {
		frames, _, _, scanErr := scanZstdFrames(prefix)
		if scanErr != nil {
			return Summary{}, scanErr
		}
		if len(frames) == 0 {
			return Summary{}, fmt.Errorf("incomplete first Zstandard frame")
		}
		// DSH writes the session header as its own frame; session/title
		// is usually a later concatenated frame in the same prefix.
		complete := prefix[:frames[len(frames)-1].end]
		dec, decErr := zstd.NewReader(bytes.NewReader(complete), zstd.WithDecoderConcurrency(1))
		if decErr != nil {
			return Summary{}, decErr
		}
		defer dec.Close()
		plain, err = io.ReadAll(dec)
		if err != nil {
			return Summary{}, err
		}
	}
	return summaryFromPrefix(plain, a)
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
