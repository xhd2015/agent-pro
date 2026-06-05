package explain

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Message string `json:"message"`
}

type RunnerMeta map[string]json.RawMessage

type SessionData struct {
	AgentRunner      string          `json:"agent_runner"`
	Model            string          `json:"model"`
	AgentRunnersMeta RunnerMeta      `json:"agent_runners_meta"`
	Messages         []Message       `json:"messages"`
}

type sessionDir struct {
	dir       string
	timestamp time.Time
	data      SessionData
}

const (
	sessionsBaseDir = ".agent-pro/dedicated-agents/explain/sessions"
	fileName        = "session.data"
)

var nonAlphanum = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func sessionsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, sessionsBaseDir), nil
}

func slugFromPrompt(prompt string) string {
	s := strings.ToLower(prompt)
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, s)
	s = nonAlphanum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}

func makeSessionName(prompt string) string {
	ts := time.Now().Format("2006-01-02-15-04-05")
	slug := slugFromPrompt(prompt)
	return ts + "-" + slug
}

func saveSession(prompt string, data SessionData) (string, error) {
	baseDir, err := sessionsDir()
	if err != nil {
		return "", err
	}
	dirName := makeSessionName(prompt)
	dir := filepath.Join(baseDir, dirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create session dir: %w", err)
	}

	dataPath := filepath.Join(dir, fileName)
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal session data: %w", err)
	}
	if err := os.WriteFile(dataPath, bytes, 0644); err != nil {
		return "", fmt.Errorf("write session data: %w", err)
	}
	return dir, nil
}

func updateSession(dir string, data SessionData) error {
	dataPath := filepath.Join(dir, fileName)
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session data: %w", err)
	}
	if err := os.WriteFile(dataPath, bytes, 0644); err != nil {
		return fmt.Errorf("write session data: %w", err)
	}
	return nil
}

func readSession(dir string) (SessionData, error) {
	dataPath := filepath.Join(dir, fileName)
	bytes, err := os.ReadFile(dataPath)
	if err != nil {
		return SessionData{}, fmt.Errorf("read session data: %w", err)
	}
	var data SessionData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return SessionData{}, fmt.Errorf("unmarshal session data: %w", err)
	}
	return data, nil
}

func listSessions() ([]sessionDir, error) {
	baseDir, err := sessionsDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read sessions dir: %w", err)
	}

	var result []sessionDir
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(baseDir, e.Name())
		data, err := readSession(dir)
		if err != nil {
			continue
		}
		ts, err := parseTimestamp(e.Name())
		if err != nil {
			continue
		}
		result = append(result, sessionDir{
			dir:       dir,
			timestamp: ts,
			data:      data,
		})
	}
	return result, nil
}

func parseTimestamp(dirName string) (time.Time, error) {
	if len(dirName) < 19 {
		return time.Time{}, fmt.Errorf("dir name too short for timestamp")
	}
	return time.Parse("2006-01-02-15-04-05", dirName[:19])
}

type MatchResult struct {
	SessionDir string
	Data       SessionData
}

func findMatchingSession(userInput string) (*MatchResult, error) {
	sessions, err := listSessions()
	if err != nil {
		return nil, err
	}

	type candidate struct {
		dir       string
		timestamp time.Time
		data      SessionData
		msgLen    int
	}

	var best *candidate
	for _, s := range sessions {
		for _, m := range s.data.Messages {
			if m.Role != "user" {
				continue
			}
			if !strings.HasPrefix(userInput, m.Message) {
				continue
			}
			msgLen := len(m.Message)
			if best == nil || msgLen > best.msgLen || (msgLen == best.msgLen && s.timestamp.After(best.timestamp)) {
				best = &candidate{
					dir:       s.dir,
					timestamp: s.timestamp,
					data:      s.data,
					msgLen:    msgLen,
				}
			}
		}
	}

	if best == nil {
		return nil, nil
	}

	return &MatchResult{
		SessionDir: best.dir,
		Data:       best.data,
	}, nil
}

func allUserMessages(data SessionData) []string {
	var msgs []string
	for _, m := range data.Messages {
		if m.Role == "user" {
			msgs = append(msgs, m.Message)
		}
	}
	return msgs
}

func sortSessionDirsByTimestamp(dirs []sessionDir) {
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].timestamp.Before(dirs[j].timestamp)
	})
}
