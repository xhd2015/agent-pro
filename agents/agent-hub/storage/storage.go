package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/agents/agent-hub/model"
)

type Store struct {
	Home string
}

func New(home string) *Store {
	return &Store{Home: home}
}

func (s *Store) cursorPath(consumerID string) string {
	safe := strings.NewReplacer("/", "__", string(os.PathSeparator), "__").Replace(consumerID)
	return filepath.Join(s.Home, "consumers", safe+".cursor.json")
}

func (s *Store) Append(event model.NormalizedEvent, receivedAt time.Time) (model.Envelope, error) {
	if receivedAt.IsZero() {
		receivedAt = time.Now().UTC()
	}
	partition := receivedAt.UTC().Format("2006-01-02")
	dir := s.partitionDir(partition)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return model.Envelope{}, err
	}
	offset, err := nextOffset(filepath.Join(dir, "events.jsonl"))
	if err != nil {
		return model.Envelope{}, err
	}
	env := model.Envelope{
		SchemaVersion: model.SchemaVersionEventV1,
		EventID:       fmt.Sprintf("%s:%d", partition, offset),
		Partition:     partition,
		Offset:        offset,
		ReceivedAt:    receivedAt.UTC(),
		Event:         event,
	}
	if err := env.Validate(); err != nil {
		return model.Envelope{}, err
	}
	line, err := json.Marshal(env)
	if err != nil {
		return model.Envelope{}, err
	}
	logPath := filepath.Join(dir, "events.jsonl")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return model.Envelope{}, err
	}
	defer f.Close()
	pos, err := f.Seek(0, 2)
	if err != nil {
		return model.Envelope{}, err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		return model.Envelope{}, err
	}
	if err := appendIndex(filepath.Join(dir, "events.idx"), offset, pos); err != nil {
		return model.Envelope{}, err
	}
	if err := s.project(env); err != nil {
		return model.Envelope{}, err
	}
	return env, nil
}

func (s *Store) ReadBatch(cursor model.Cursor, limit int) ([]model.Envelope, model.Cursor, bool, error) {
	if limit <= 0 {
		return nil, cursor, false, fmt.Errorf("limit must be > 0")
	}
	partitions, err := s.Partitions()
	if err != nil {
		return nil, cursor, false, err
	}
	if len(partitions) == 0 {
		return nil, cursor, false, nil
	}
	idx := sort.SearchStrings(partitions, cursor.Partition)
	if idx >= len(partitions) {
		return nil, cursor, false, nil
	}
	if idx < len(partitions) && partitions[idx] != cursor.Partition {
		cursor = model.Cursor{Partition: partitions[idx], Offset: 0}
	}
	var out []model.Envelope
	next := cursor
	for idx < len(partitions) && len(out) < limit {
		partition := partitions[idx]
		events, err := s.readPartition(partition)
		if err != nil {
			return nil, cursor, false, err
		}
		start := int64(0)
		if partition == next.Partition {
			start = next.Offset
		}
		for start < int64(len(events)) && len(out) < limit {
			out = append(out, events[start])
			start++
		}
		if start < int64(len(events)) {
			next = model.Cursor{Partition: partition, Offset: start}
			return out, next, true, nil
		}
		idx++
		if idx < len(partitions) {
			next = model.Cursor{Partition: partitions[idx], Offset: 0}
		} else {
			next = model.Cursor{Partition: partition, Offset: int64(len(events))}
		}
	}
	hasMore := false
	if len(out) == limit {
		hasMore = s.hasEventAt(next)
	}
	return out, next, hasMore, nil
}

func (s *Store) Fetch(consumerID string, limit int, peek bool) (model.FetchResponse, error) {
	cur, err := s.LoadCursor(consumerID)
	if err != nil {
		if !os.IsNotExist(err) {
			return model.FetchResponse{}, err
		}
		partitions, err := s.Partitions()
		if err != nil {
			return model.FetchResponse{}, err
		}
		if len(partitions) > 0 {
			cur = model.Cursor{Partition: partitions[0], Offset: 0}
		}
	}
	events, next, hasMore, err := s.ReadBatch(cur, limit)
	if err != nil {
		return model.FetchResponse{}, err
	}
	if !peek {
		if cur.Partition != "" || next.Partition != "" {
			if err := s.SaveCursor(consumerID, next); err != nil {
				return model.FetchResponse{}, err
			}
		}
	} else {
		next = cur
	}
	if events == nil {
		events = []model.Envelope{}
	}
	return model.FetchResponse{
		ConsumerID:     consumerID,
		Events:         events,
		PreviousCursor: cur,
		NextCursor:     next,
		HasMore:        hasMore,
	}, nil
}

func (s *Store) SaveCursor(consumerID string, cursor model.Cursor) error {
	if strings.TrimSpace(consumerID) == "" {
		return fmt.Errorf("consumer id is required")
	}
	if err := cursor.Validate(); err != nil {
		return err
	}
	path := s.cursorPath(consumerID)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(struct {
		ConsumerID string       `json:"consumer_id"`
		Cursor     model.Cursor `json:"cursor"`
		UpdatedAt  time.Time    `json:"updated_at"`
	}{ConsumerID: consumerID, Cursor: cursor, UpdatedAt: time.Now().UTC()}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *Store) LoadCursor(consumerID string) (model.Cursor, error) {
	data, err := os.ReadFile(s.cursorPath(consumerID))
	if err != nil {
		return model.Cursor{}, err
	}
	var doc struct {
		Cursor model.Cursor `json:"cursor"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return model.Cursor{}, err
	}
	return doc.Cursor, doc.Cursor.Validate()
}

func (s *Store) Partitions() ([]string, error) {
	root := filepath.Join(s.Home, "events")
	var partitions []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != "events.jsonl" {
			return nil
		}
		rel, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) == 3 {
			partitions = append(partitions, parts[0]+"-"+parts[1]+"-"+parts[2])
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	sort.Strings(partitions)
	return partitions, nil
}

func (s *Store) RebuildIndexes() error {
	partitions, err := s.Partitions()
	if err != nil {
		return err
	}
	for _, partition := range partitions {
		if err := s.rebuildIndex(partition); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) RebuildSessions() error {
	if err := os.RemoveAll(filepath.Join(s.Home, "sessions")); err != nil {
		return err
	}
	partitions, err := s.Partitions()
	if err != nil {
		return err
	}
	for _, partition := range partitions {
		events, err := s.readPartition(partition)
		if err != nil {
			return err
		}
		for _, env := range events {
			if err := s.project(env); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) partitionDir(partition string) string {
	return filepath.Join(s.Home, "events", partition[0:4], partition[5:7], partition[8:10])
}

func (s *Store) readPartition(partition string) ([]model.Envelope, error) {
	path := filepath.Join(s.partitionDir(partition), "events.jsonl")
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var events []model.Envelope
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		var env model.Envelope
		if err := json.Unmarshal(scanner.Bytes(), &env); err != nil {
			return nil, err
		}
		events = append(events, env)
	}
	return events, scanner.Err()
}

func (s *Store) hasEventAt(cursor model.Cursor) bool {
	events, err := s.readPartition(cursor.Partition)
	return err == nil && cursor.Offset < int64(len(events))
}

func (s *Store) rebuildIndex(partition string) error {
	path := filepath.Join(s.partitionDir(partition), "events.jsonl")
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	idxPath := filepath.Join(s.partitionDir(partition), "events.idx")
	idx, err := os.Create(idxPath)
	if err != nil {
		return err
	}
	defer idx.Close()
	var offset int64
	var pos int64
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		if err := writeIndex(idx, offset, pos); err != nil {
			return err
		}
		pos += int64(len(scanner.Bytes())) + 1
		offset++
	}
	return scanner.Err()
}

func (s *Store) project(env model.Envelope) error {
	sessionID := strings.TrimSpace(env.Event.RunnerSessionID)
	if sessionID == "" {
		return nil
	}
	status := ""
	switch env.Event.EventType {
	case model.EventSessionStarted:
		status = "running"
	case model.EventSessionFinished:
		status = "completed"
	case model.EventSessionFailed:
		status = "failed"
	default:
		return nil
	}
	runner := sanitize(env.Event.Runner)
	sid := sanitize(sessionID)
	targetDir := filepath.Join(s.Home, "sessions", runner, sid)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}
	existing, _ := s.GetSession(runner, sessionID)
	if existing != nil && existing.Status == "completed" && status == "running" {
	} else if existing != nil && existing.Status == "failed" && status == "running" {
	} else {
	}
	data := model.SessionData{
		Runner:          env.Event.Runner,
		RunnerSessionID: sessionID,
		Status:          status,
		LastEvent:       &model.Cursor{Partition: env.Partition, Offset: env.Offset},
	}
	return s.writeSessionFile(runner, sessionID, data)
}

func (s *Store) sessionDir(runner, sessionID string) string {
	return filepath.Join(s.Home, "sessions", sanitize(runner), sanitize(sessionID))
}

func (s *Store) sessionFilePath(runner, sessionID string) string {
	return filepath.Join(s.sessionDir(runner, sessionID), "session.json")
}

func (s *Store) writeSessionFile(runner, sessionID string, data model.SessionData) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	dir := s.sessionDir(runner, sessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(s.sessionFilePath(runner, sessionID), raw, 0644)
}

func (s *Store) GetSession(runner, sessionID string) (*model.SessionData, error) {
	path := s.sessionFilePath(runner, sessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var sd model.SessionData
	if err := json.Unmarshal(data, &sd); err != nil {
		return nil, err
	}
	return &sd, nil
}

func (s *Store) WriteSession(runner, sessionID string, data model.SessionData) error {
	return s.writeSessionFile(runner, sessionID, data)
}

func (s *Store) messagesPath(runner, sessionID string) string {
	return filepath.Join(s.sessionDir(runner, sessionID), "messages.jsonl")
}

func (s *Store) GetMessages(runner, sessionID string) ([]model.Message, error) {
	path := s.messagesPath(runner, sessionID)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var msgs []model.Message
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var msg model.Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, scanner.Err()
}

func (s *Store) AppendMessage(runner, sessionID string, msg model.Message) error {
	dir := s.sessionDir(runner, sessionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := s.messagesPath(runner, sessionID)
	line, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(line, '\n'))
	return err
}

func (s *Store) ClearMessages(runner, sessionID string) error {
	path := s.messagesPath(runner, sessionID)
	if err := os.WriteFile(path, nil, 0644); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}

func (s *Store) GetAndClearMessages(runner, sessionID string) ([]model.Message, error) {
	msgs, err := s.GetMessages(runner, sessionID)
	if err != nil {
		return nil, err
	}
	if err := s.ClearMessages(runner, sessionID); err != nil {
		return msgs, err
	}
	return msgs, nil
}

func nextOffset(path string) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	defer f.Close()
	var offset int64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		offset++
	}
	return offset, scanner.Err()
}

func appendIndex(path string, offset int64, bytePos int64) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return writeIndex(f, offset, bytePos)
}

func writeIndex(f *os.File, offset int64, bytePos int64) error {
	data, err := json.Marshal(struct {
		Offset int64 `json:"offset"`
		Byte   int64 `json:"byte"`
	}{Offset: offset, Byte: bytePos})
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	return err
}

func sanitize(name string) string {
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_")
	return replacer.Replace(name)
}
