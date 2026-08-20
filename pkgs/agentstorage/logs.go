package agentstorage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LogRecord is one durable, session-scoped runtime diagnostic. Logs are kept
// separate from session state so normal progress never pollutes meta.json.
type LogRecord struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Component string `json:"component"`
	Message   string `json:"message"`
}

// LogsPath is the durable JSONL diagnostic path for a session.
func LogsPath(home, sessionID string) string {
	return filepath.Join(home, "sessions", sessionID, "logs.jsonl")
}

// AppendErrorLog records a runtime error without ever using an attached TTY as
// a diagnostic sink. The session must already exist.
func AppendErrorLog(home, sessionID, component string, err error) error {
	if strings.TrimSpace(sessionID) == "" || err == nil {
		return nil
	}
	dir := filepath.Dir(LogsPath(home, sessionID))
	if _, statErr := os.Stat(filepath.Join(dir, "meta.json")); statErr != nil {
		if os.IsNotExist(statErr) {
			return fmt.Errorf("session not found: %s", sessionID)
		}
		return statErr
	}
	record := LogRecord{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     "error",
		Component: strings.TrimSpace(component),
		Message:   err.Error(),
	}
	data, marshalErr := json.Marshal(record)
	if marshalErr != nil {
		return marshalErr
	}
	f, openErr := os.OpenFile(LogsPath(home, sessionID), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if openErr != nil {
		return openErr
	}
	defer f.Close()
	_, writeErr := f.Write(append(data, '\n'))
	return writeErr
}

// ReadLogs returns parsed session logs and the original JSONL bytes. A session
// with no errors has no log file and returns an empty result.
func ReadLogs(home, sessionID string) ([]LogRecord, []byte, error) {
	data, err := os.ReadFile(LogsPath(home, sessionID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	var records []LogRecord
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var record LogRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, nil, fmt.Errorf("parse logs.jsonl: %w", err)
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	return records, data, nil
}
