package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Durable conversation map + per-session message log under
// $HOME/.agent-pro/slack-local-bot/.

const sessionsMapFileName = "sessions.json"

type durableSessionEntry struct {
	SessionID          string `json:"session_id"`
	ChannelID          string `json:"channel_id"`
	ThreadTS           string `json:"thread_ts"`
	ConfigPath         string `json:"config_path"`
	Kind               string `json:"kind"`
	ReplyMode          string `json:"reply_mode"`
	CreatedAt          string `json:"created_at,omitempty"`
	UpdatedAt          string `json:"updated_at,omitempty"`
	LastMessagePreview string `json:"last_message_preview,omitempty"`
}

type sessionsMapFile struct {
	Version int                   `json:"version"`
	Entries []durableSessionEntry `json:"entries"`
}

type sessionLogEntry struct {
	MessageID string `json:"message_id"`
	TS        string `json:"ts"`
	User      string `json:"user"`
	Text      string `json:"text"`
	Direction string `json:"direction"` // in | out
}

func sessionsMapPath(dataRoot string) string {
	return filepath.Join(dataRoot, sessionsMapFileName)
}

func sessionMessagesJSONLPath(dataRoot, sessionID string) string {
	return filepath.Join(dataRoot, "sessions", sessionID, "messages.jsonl")
}

func loadSessionsMap(dataRoot string) (*sessionsMapFile, error) {
	path := sessionsMapPath(dataRoot)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &sessionsMapFile{Version: 1, Entries: []durableSessionEntry{}}, nil
		}
		return nil, err
	}
	var doc sessionsMapFile
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if doc.Version == 0 {
		doc.Version = 1
	}
	if doc.Entries == nil {
		doc.Entries = []durableSessionEntry{}
	}
	return &doc, nil
}

func saveSessionsMap(dataRoot string, doc *sessionsMapFile) error {
	if doc == nil {
		doc = &sessionsMapFile{Version: 1}
	}
	if doc.Version == 0 {
		doc.Version = 1
	}
	path := sessionsMapPath(dataRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func lookupSessionEntry(dataRoot, sessionID string) (*durableSessionEntry, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("session id required")
	}
	doc, err := loadSessionsMap(dataRoot)
	if err != nil {
		return nil, err
	}
	for i := range doc.Entries {
		if doc.Entries[i].SessionID == sessionID {
			entry := doc.Entries[i]
			return &entry, nil
		}
	}
	return nil, fmt.Errorf("session not found: %s", sessionID)
}

// upsertSessionEntry creates or updates a map entry for the given session.
func upsertSessionEntry(dataRoot string, entry durableSessionEntry) error {
	entry.SessionID = strings.TrimSpace(entry.SessionID)
	if entry.SessionID == "" {
		return fmt.Errorf("session id required")
	}
	if entry.Kind == "" {
		entry.Kind = "channel"
	}
	if entry.ReplyMode == "" {
		entry.ReplyMode = "channel"
	}
	now := time.Now().UTC().Format(time.RFC3339)
	doc, err := loadSessionsMap(dataRoot)
	if err != nil {
		return err
	}
	found := false
	for i := range doc.Entries {
		if doc.Entries[i].SessionID != entry.SessionID {
			continue
		}
		// Preserve created_at if present.
		if doc.Entries[i].CreatedAt != "" {
			entry.CreatedAt = doc.Entries[i].CreatedAt
		} else if entry.CreatedAt == "" {
			entry.CreatedAt = now
		}
		entry.UpdatedAt = now
		doc.Entries[i] = entry
		found = true
		break
	}
	if !found {
		if entry.CreatedAt == "" {
			entry.CreatedAt = now
		}
		entry.UpdatedAt = now
		doc.Entries = append(doc.Entries, entry)
	}
	return saveSessionsMap(dataRoot, doc)
}

func appendSessionMessage(dataRoot, sessionID string, msg sessionLogEntry) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return fmt.Errorf("session id required")
	}
	path := sessionMessagesJSONLPath(dataRoot, sessionID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if msg.MessageID == "" {
		msg.MessageID = msg.TS
	}
	line, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(append(line, '\n')); err != nil {
		return err
	}
	return nil
}

func readSessionMessages(dataRoot, sessionID string) ([]sessionLogEntry, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, fmt.Errorf("session id required")
	}
	path := sessionMessagesJSONLPath(dataRoot, sessionID)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []sessionLogEntry{}, nil
		}
		return nil, err
	}
	defer f.Close()

	var out []sessionLogEntry
	sc := bufio.NewScanner(f)
	// Allow long message lines.
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var m sessionLogEntry
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			return nil, fmt.Errorf("parse messages.jsonl: %w", err)
		}
		out = append(out, m)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// filterMessagesAfterID returns messages after the first occurrence of message_id
// (or ts) equal to afterID, in log order. If afterID is empty, returns all.
func filterMessagesAfterID(msgs []sessionLogEntry, afterID string) []sessionLogEntry {
	afterID = strings.TrimSpace(afterID)
	if afterID == "" {
		return msgs
	}
	for i, m := range msgs {
		id := m.MessageID
		if id == "" {
			id = m.TS
		}
		if id == afterID {
			if i+1 >= len(msgs) {
				return []sessionLogEntry{}
			}
			return msgs[i+1:]
		}
	}
	// afterID not found: return empty (strict after filter).
	return []sessionLogEntry{}
}
