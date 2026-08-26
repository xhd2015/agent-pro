package explain

import (
	"crypto/sha256"
	"encoding/hex"
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
	AgentRunner      string     `json:"agent_runner"`
	Model            string     `json:"model"`
	AgentRunnersMeta RunnerMeta `json:"agent_runners_meta"`
	Messages         []Message  `json:"messages"`
}

type sessionDir struct {
	dir       string
	timestamp time.Time
	data      SessionData
}

const (
	defaultSessionsBaseDir = ".agent-pro/dedicated-agents/explain"
	debugConfigHomeEnv     = "AGENT_PRO_DEDICATED_AGENT_EXPLAIN_DEBUG_CONFIG_HOME"
	fileName               = "session.data"
)

var nonAlphanum = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func sessionsDir() (string, error) {
	root, err := explainRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "sessions"), nil
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
	hash := sha256.Sum256([]byte(prompt))
	hashStr := hex.EncodeToString(hash[:])
	return ts + "-" + slug + "-" + hashStr[:8]
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

// ListRecentSessions returns sessions sorted by dirname timestamp descending,
// limited to at most limit entries, plus the total count before limiting.
// limit is normalized the same way as the list CLI: <=0 → default 10, cap 100.
func ListRecentSessions(limit int) (sessions []sessionDir, total int, err error) {
	sessions, err = listSessions()
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].timestamp.After(sessions[j].timestamp)
	})
	total = len(sessions)
	limit = normalizeListLimit(limit)
	if len(sessions) > limit {
		sessions = sessions[:limit]
	}
	return sessions, total, nil
}

func parseTimestamp(dirName string) (time.Time, error) {
	if len(dirName) < 19 {
		return time.Time{}, fmt.Errorf("dir name too short for timestamp")
	}
	return time.Parse("2006-01-02-15-04-05", dirName[:19])
}

type MatchResult struct {
	SessionDir   string
	Data         SessionData
	MatchedCount int
}

func findMatchingSession(args []string) (*MatchResult, error) {
	sessions, err := listSessions()
	if err != nil {
		return nil, err
	}

	type candidate struct {
		dir       string
		timestamp time.Time
		data      SessionData
		n         int
	}

	var best *candidate
	for _, s := range sessions {
		userMsgs := userMessageSlice(s.data)
		n := countPrefixMatch(userMsgs, args)
		if n == 0 {
			continue
		}
		if best == nil || n > best.n || (n == best.n && s.timestamp.After(best.timestamp)) {
			best = &candidate{
				dir:       s.dir,
				timestamp: s.timestamp,
				data:      s.data,
				n:         n,
			}
		}
	}

	if best == nil {
		return nil, nil
	}

	return &MatchResult{
		SessionDir:   best.dir,
		Data:         best.data,
		MatchedCount: best.n,
	}, nil
}

func countPrefixMatch(stored []string, input []string) int {
	n := 0
	for n < len(stored) && n < len(input) {
		if stored[n] != input[n] {
			break
		}
		n++
	}
	if n > 0 && n < len(input) {
		return n
	}
	return 0
}

func userMessageSlice(data SessionData) []string {
	var msgs []string
	for _, m := range data.Messages {
		if m.Role == "user" {
			msgs = append(msgs, m.Message)
		}
	}
	return msgs
}
