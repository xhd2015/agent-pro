package agentstorage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	types "github.com/xhd2015/agent-pro/agent/event/types"
)

type fileStore struct {
	home string
}

func osMkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

func (s *fileStore) Home() string {
	return s.home
}

func (s *fileStore) configPath() string {
	return filepath.Join(s.home, "config.json")
}

func (s *fileStore) sessionDir(runner, sessionID string) string {
	return filepath.Join(s.home, "sessions", runner, sessionID)
}

func (s *fileStore) Config() (Config, error) {
	data, err := os.ReadFile(s.configPath())
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, err
	}
	var cfg Config
	if len(strings.TrimSpace(string(data))) == 0 {
		return Config{}, nil
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (s *fileStore) SaveConfig(cfg Config) error {
	if err := osMkdirAll(s.home); err != nil {
		return err
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(s.configPath(), data, 0644)
}

func (s *fileStore) CreateSession(runner, sessionID string, meta SessionMeta) error {
	dir := s.sessionDir(runner, sessionID)
	if err := osMkdirAll(dir); err != nil {
		return err
	}
	now := nowRFC3339()
	meta.Runner = runner
	meta.SessionID = sessionID
	if meta.CreatedAt == "" {
		meta.CreatedAt = now
	}
	meta.UpdatedAt = now
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "meta.json"), data, 0644)
}

func (s *fileStore) GetSession(runner, sessionID string) (*Session, error) {
	path := filepath.Join(s.sessionDir(runner, sessionID), "meta.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session not found: %s/%s", runner, sessionID)
		}
		return nil, err
	}
	var meta SessionMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &Session{Meta: meta}, nil
}

func (s *fileStore) UpdateSessionStatus(runner, sessionID, status string) error {
	sess, err := s.GetSession(runner, sessionID)
	if err != nil {
		return err
	}
	sess.Meta.Status = status
	sess.Meta.UpdatedAt = nowRFC3339()
	data, err := json.Marshal(sess.Meta)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.sessionDir(runner, sessionID), "meta.json"), data, 0644)
}

func (s *fileStore) UpdateSessionRunnerSessionID(runner, sessionID, runnerSessionID string) error {
	sess, err := s.GetSession(runner, sessionID)
	if err != nil {
		return err
	}
	runnerSessionID = strings.TrimSpace(runnerSessionID)
	if runnerSessionID == "" {
		return nil
	}
	sess.Meta.RunnerSessionID = runnerSessionID
	sess.Meta.UpdatedAt = nowRFC3339()
	data, err := json.Marshal(sess.Meta)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.sessionDir(runner, sessionID), "meta.json"), data, 0644)
}

func (s *fileStore) ListSessions(runner string) ([]SessionMeta, error) {
	root := filepath.Join(s.home, "sessions", runner)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []SessionMeta
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		sess, err := s.GetSession(runner, ent.Name())
		if err != nil {
			continue
		}
		out = append(out, sess.Meta)
	}
	return out, nil
}

func (s *fileStore) eventsPath(runner, sessionID string) string {
	return filepath.Join(s.sessionDir(runner, sessionID), "events.jsonl")
}

func (s *fileStore) AppendEvent(runner, sessionID string, ev types.AgentEvent) error {
	dir := s.sessionDir(runner, sessionID)
	if err := osMkdirAll(dir); err != nil {
		return err
	}
	line, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(s.eventsPath(runner, sessionID), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(line, '\n'))
	return err
}

func (s *fileStore) ReadEvents(runner, sessionID string, afterOffset int64) ([]types.AgentEvent, int64, error) {
	path := s.eventsPath(runner, sessionID)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []types.AgentEvent{}, 0, nil
		}
		return nil, 0, err
	}
	defer f.Close()

	if afterOffset > 0 {
		if _, err := f.Seek(afterOffset, io.SeekStart); err != nil {
			return nil, 0, err
		}
	}

	var events []types.AgentEvent
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimSpace(line)
			if line != "" {
				var ev types.AgentEvent
				if err := json.Unmarshal([]byte(line), &ev); err != nil {
					return nil, 0, err
				}
				events = append(events, ev)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, 0, err
		}
	}

	offset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, 0, err
	}
	return events, offset, nil
}

func (s *fileStore) messagesPath(runner, sessionID string) string {
	return filepath.Join(s.sessionDir(runner, sessionID), "messages.jsonl")
}

func (s *fileStore) readAllMessages(runner, sessionID string) ([]Message, error) {
	path := s.messagesPath(runner, sessionID)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Message{}, nil
		}
		return nil, err
	}
	var msgs []Message
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var m Message
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return msgs, nil
}

func (s *fileStore) writeAllMessages(runner, sessionID string, msgs []Message) error {
	dir := s.sessionDir(runner, sessionID)
	if err := osMkdirAll(dir); err != nil {
		return err
	}
	path := s.messagesPath(runner, sessionID)
	if len(msgs) == 0 {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	var b strings.Builder
	for _, m := range msgs {
		line, err := json.Marshal(m)
		if err != nil {
			return err
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func (s *fileStore) AppendMessage(runner, sessionID, text string) (Message, error) {
	msgs, err := s.readAllMessages(runner, sessionID)
	if err != nil {
		return Message{}, err
	}
	msg := Message{
		ID:        fmt.Sprintf("msg_%d", len(msgs)+1),
		Text:      text,
		SessionID: sessionID,
		CreatedAt: nowRFC3339(),
	}
	msgs = append(msgs, msg)
	if err := s.writeAllMessages(runner, sessionID, msgs); err != nil {
		return Message{}, err
	}
	return msg, nil
}

func (s *fileStore) ListMessages(runner, sessionID string) ([]Message, error) {
	msgs, err := s.readAllMessages(runner, sessionID)
	if err != nil {
		return nil, err
	}
	if msgs == nil {
		return []Message{}, nil
	}
	return msgs, nil
}

func (s *fileStore) PopMessages(runner, sessionID string) ([]Message, error) {
	msgs, err := s.readAllMessages(runner, sessionID)
	if err != nil {
		return nil, err
	}
	if msgs == nil {
		msgs = []Message{}
	}
	if err := s.writeAllMessages(runner, sessionID, nil); err != nil {
		return nil, err
	}
	return msgs, nil
}