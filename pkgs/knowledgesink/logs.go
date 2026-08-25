package knowledgesink

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LogsFileName is the unified session debug transcript (append-only JSONL).
const LogsFileName = "logs.jsonl"

// LogLine is one JSON object per line in logs.jsonl.
type LogLine struct {
	TS        string `json:"ts"`
	Stream    string `json:"stream"` // stdout | stderr
	Text      string `json:"text"`
	SinkIndex *int   `json:"sink_index,omitempty"`
	Trigger   string `json:"trigger,omitempty"` // ui | auto | cli
}

func LogsPath(sessionDir string) string {
	return filepath.Join(sessionDir, LogsFileName)
}

// AppendLog appends one JSONL record. Best-effort; errors are ignored by callers
// that also tee to a live UI log.
func AppendLog(sessionDir, stream, text, trigger string, sinkIndex int, now time.Time) error {
	sessionDir = strings.TrimSpace(sessionDir)
	if sessionDir == "" || text == "" {
		return nil
	}
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		return err
	}
	if now.IsZero() {
		now = time.Now()
	}
	line := LogLine{
		TS:      FormatTime(now),
		Stream:  strings.TrimSpace(stream),
		Text:    text,
		Trigger: strings.TrimSpace(trigger),
	}
	if sinkIndex >= 0 {
		idx := sinkIndex
		line.SinkIndex = &idx
	}
	b, err := json.Marshal(line)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	f, err := os.OpenFile(LogsPath(sessionDir), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(b)
	return err
}
