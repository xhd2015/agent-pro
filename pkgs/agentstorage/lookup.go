package agentstorage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	errEmptyGrokSessionID = "--grok-session-id requires a non-empty value"
)

// runnerSessionIndexEntry is one mapping in index/by-runner-session/<uuid>.json.
type runnerSessionIndexEntry struct {
	SessionID string `json:"session_id"`
	Runner    string `json:"runner"`
}

// IsGrokRunner reports whether runner is exactly trimmed "grok" or "grok-tty".
func IsGrokRunner(runner string) bool {
	r := strings.TrimSpace(runner)
	return r == "grok" || r == "grok-tty"
}

// IsCodexRunner reports whether runner is exactly trimmed "codex" or "codex-tty".
func IsCodexRunner(runner string) bool {
	r := strings.TrimSpace(runner)
	return r == "codex" || r == "codex-tty"
}

// ListByRunnerSessionID returns all session metas whose trimmed runner_session_id
// equals trimmed id. When runners is non-empty, meta.runner must be one of those
// exact trimmed names. Empty/whitespace id returns errEmptyGrokSessionID.
func ListByRunnerSessionID(store Store, id string, runners ...string) ([]SessionMeta, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("%s", errEmptyGrokSessionID)
	}
	filter := makeRunnerFilter(runners)
	entries, err := lookupRunnerSessionEntries(store, id)
	if err != nil {
		return nil, err
	}
	var out []SessionMeta
	for _, e := range entries {
		if filter != nil && !filter[strings.TrimSpace(e.Runner)] {
			continue
		}
		sess, gerr := store.GetSession(e.SessionID)
		if gerr != nil {
			continue
		}
		out = append(out, sess.Meta)
	}
	return out, nil
}

// FindByGrokSessionID finds the unique grok/grok-tty session for runner_session_id.
// Cardinality: 0 → not found; 1 → that meta; 2+ → ambiguous (session ids ascending).
func FindByGrokSessionID(store Store, id string) (SessionMeta, error) {
	return findByRunnerSessionID(store, id, "grok", []string{"grok", "grok-tty"})
}

// FindByCodexSessionID finds the unique codex/codex-tty session for runner_session_id.
// Cardinality: 0 → not found; 1 → that meta; 2+ → ambiguous (session ids ascending).
func FindByCodexSessionID(store Store, id string) (SessionMeta, error) {
	return findByRunnerSessionID(store, id, "codex", []string{"codex", "codex-tty"})
}

func findByRunnerSessionID(store Store, id, label string, runners []string) (SessionMeta, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return SessionMeta{}, fmt.Errorf("%s", errEmptyGrokSessionID)
	}
	matches, err := ListByRunnerSessionID(store, id, runners...)
	if err != nil {
		return SessionMeta{}, err
	}
	switch len(matches) {
	case 0:
		return SessionMeta{}, fmt.Errorf("session not found: no %s session with runner_session_id %s", label, id)
	case 1:
		return matches[0], nil
	default:
		ids := make([]string, 0, len(matches))
		for _, m := range matches {
			ids = append(ids, m.SessionID)
		}
		sort.Strings(ids)
		return SessionMeta{}, fmt.Errorf("ambiguous %s-session-id %s: multiple matches: %s", label, id, strings.Join(ids, ", "))
	}
}

func makeRunnerFilter(runners []string) map[string]bool {
	if len(runners) == 0 {
		return nil
	}
	filter := make(map[string]bool, len(runners))
	for _, r := range runners {
		filter[strings.TrimSpace(r)] = true
	}
	return filter
}

func lookupRunnerSessionEntries(store Store, id string) ([]runnerSessionIndexEntry, error) {
	home := store.Home()
	if runnerSessionIndexWarm(home) {
		return readRunnerSessionIndexFile(home, id)
	}
	byUUID, err := rebuildRunnerSessionIndex(store)
	if err != nil {
		return nil, err
	}
	return byUUID[id], nil
}

func indexDir(home string) string {
	return filepath.Join(home, "index")
}

func byRunnerSessionDir(home string) string {
	return filepath.Join(indexDir(home), "by-runner-session")
}

func generationPath(home string) string {
	return filepath.Join(indexDir(home), "generation")
}

func byRunnerGenPath(home string) string {
	return filepath.Join(byRunnerSessionDir(home), ".gen")
}

func runnerSessionIndexFile(home, uuid string) string {
	return filepath.Join(byRunnerSessionDir(home), uuid+".json")
}

func readGenerationFile(path string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(string(data)), true
}

// runnerSessionIndexWarm is true only when generation and .gen both exist and match.
func runnerSessionIndexWarm(home string) bool {
	gen, genOK := readGenerationFile(generationPath(home))
	if !genOK {
		return false
	}
	byGen, byOK := readGenerationFile(byRunnerGenPath(home))
	if !byOK {
		return false
	}
	return gen == byGen
}

func readRunnerSessionIndexFile(home, uuid string) ([]runnerSessionIndexEntry, error) {
	path := runnerSessionIndexFile(home, uuid)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, nil
	}
	var entries []runnerSessionIndexEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func rebuildRunnerSessionIndex(store Store) (map[string][]runnerSessionIndexEntry, error) {
	home := store.Home()
	list, err := store.ListSessions()
	if err != nil {
		return nil, err
	}
	byUUID := make(map[string][]runnerSessionIndexEntry)
	for _, m := range list {
		rsid := strings.TrimSpace(m.RunnerSessionID)
		if rsid == "" {
			continue
		}
		byUUID[rsid] = append(byUUID[rsid], runnerSessionIndexEntry{
			SessionID: m.SessionID,
			Runner:    m.Runner,
		})
	}

	byDir := byRunnerSessionDir(home)
	if err := os.RemoveAll(byDir); err != nil {
		return nil, err
	}
	if err := osMkdirAll(byDir); err != nil {
		return nil, err
	}

	for uuid, entries := range byUUID {
		data, err := json.Marshal(entries)
		if err != nil {
			return nil, err
		}
		if err := atomicWriteFile(runnerSessionIndexFile(home, uuid), data); err != nil {
			return nil, err
		}
	}

	gen, genOK := readGenerationFile(generationPath(home))
	if !genOK {
		gen = "0"
	}
	if err := atomicWriteFile(byRunnerGenPath(home), []byte(gen)); err != nil {
		return nil, err
	}
	return byUUID, nil
}

func atomicWriteFile(path string, data []byte) error {
	if err := osMkdirAll(filepath.Dir(path)); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func (s *fileStore) bumpGeneration() error {
	if err := osMkdirAll(indexDir(s.home)); err != nil {
		return err
	}
	path := generationPath(s.home)
	cur := int64(0)
	if text, ok := readGenerationFile(path); ok && text != "" {
		n, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return fmt.Errorf("index/generation: %w", err)
		}
		cur = n
	}
	cur++
	return atomicWriteFile(path, []byte(strconv.FormatInt(cur, 10)))
}

func (s *fileStore) clearRunnerSessionIndex() error {
	return os.RemoveAll(byRunnerSessionDir(s.home))
}
