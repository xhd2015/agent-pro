package sessions

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"time"
)

// NewestTimestamp returns the latest line-level "timestamp" in a Codex rollout
// JSONL file. Zero time if the file is missing, empty, or has no parseable stamps.
// This is the activity watermark for sink cursor / sinkability (not StartedAt, not mtime).
func NewestTimestamp(rolloutPath string) (time.Time, error) {
	rolloutPath = strings.TrimSpace(rolloutPath)
	if rolloutPath == "" {
		return time.Time{}, nil
	}
	f, err := os.Open(rolloutPath)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, nil
		}
		return time.Time{}, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	// Rollout lines can be large (encrypted reasoning payloads).
	buf := make([]byte, 0, 1024*1024)
	sc.Buffer(buf, 16*1024*1024)

	var tip time.Time
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var row struct {
			Timestamp string `json:"timestamp"`
		}
		if json.Unmarshal([]byte(line), &row) != nil {
			continue
		}
		t, perr := parseTimestamp(row.Timestamp)
		if perr != nil {
			continue
		}
		if tip.IsZero() || t.After(tip) {
			tip = t
		}
	}
	if err := sc.Err(); err != nil {
		return tip, err
	}
	return tip, nil
}

// TipForSession resolves the rollout via Find and returns NewestTimestamp.
func TipForSession(codexHome, sessionID string) (time.Time, error) {
	path, err := Find(codexHome, sessionID)
	if err != nil {
		return time.Time{}, err
	}
	return NewestTimestamp(path)
}
