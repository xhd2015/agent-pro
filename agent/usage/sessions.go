package usage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	codexmodels "github.com/xhd2015/agent-pro/agent/codex/models"
	"github.com/xhd2015/agent-pro/agent/commandcode"
	grokmodels "github.com/xhd2015/agent-pro/agent/grok/models"
)

// countSessions counts the provider's sessions on disk. It never fails the
// record: an unreadable tree is reported through SessionsBlock.Error.
func countSessions(id ProviderID, opts CollectOptions) SessionsBlock {
	var (
		block SessionsBlock
		err   error
	)
	switch id {
	case Grok:
		block, err = CountGrokSessions(opts.GrokHome)
	case Codex:
		block, err = CountCodexSessions(opts.CodexHome)
	case CommandCode:
		block, err = CountCommandCodeSessions(opts.CommandCodeHome)
	default:
		err = fmt.Errorf("unknown usage provider %q", id)
	}
	if err != nil {
		block.Error = err.Error()
	}
	return block
}

// CountGrokSessions counts grok sessions under <home>/sessions. Every session
// directory counts, including subagent and fork children. Oldest is the
// earliest summary created_at, Newest the latest summary updated_at.
func CountGrokSessions(home string) (SessionsBlock, error) {
	root := filepath.Join(resolveHome(home, grokmodels.DefaultHome()), "sessions")
	var block SessionsBlock
	err := walkFiles(root, func(path string, entry fs.DirEntry) error {
		if entry.Name() != "summary.json" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		var summary struct {
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
		}
		if err := json.Unmarshal(raw, &summary); err != nil {
			return nil
		}
		block.Total++
		block.Observe(parseTimestamp(summary.CreatedAt), parseTimestamp(summary.UpdatedAt))
		return nil
	})
	if err != nil {
		return block, err
	}
	return block, nil
}

// rolloutTimestampRe matches the local-time stamp codex puts in rollout file
// names: rollout-2026-04-27T15-05-53-<uuid>.jsonl.
var rolloutTimestampRe = regexp.MustCompile(`^rollout-(\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2})-`)

// CountCodexSessions counts codex rollout files under <home>/sessions. Oldest
// comes from the rollout file name (the session start, recorded in local time)
// and Newest from the file modification time.
func CountCodexSessions(home string) (SessionsBlock, error) {
	root := filepath.Join(resolveHome(home, codexmodels.DefaultHome()), "sessions")
	var block SessionsBlock
	err := walkFiles(root, func(path string, entry fs.DirEntry) error {
		base := entry.Name()
		if !strings.HasPrefix(base, "rollout-") || !strings.HasSuffix(base, ".jsonl") {
			return nil
		}
		block.Total++
		var started time.Time
		if match := rolloutTimestampRe.FindStringSubmatch(base); match != nil {
			if parsed, err := time.ParseInLocation("2006-01-02T15-04-05", match[1], time.Local); err == nil {
				started = parsed
			}
		}
		block.Observe(started, entryModTime(entry))
		return nil
	})
	if err != nil {
		return block, err
	}
	return block, nil
}

// commandCodeSessionRe matches the transcript file name for one session,
// excluding the .checkpoints.jsonl sidecars.
var commandCodeSessionRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\.jsonl$`)

// CountCommandCodeSessions counts Command Code transcripts under
// <home>/projects. Oldest is the earliest session timestamp in a transcript's
// first record and Newest the latest file modification time.
func CountCommandCodeSessions(home string) (SessionsBlock, error) {
	dir := strings.TrimSpace(home)
	if dir == "" {
		dir = commandcode.DefaultHome()
	}
	root := filepath.Join(dir, "projects")
	var block SessionsBlock
	err := walkFiles(root, func(path string, entry fs.DirEntry) error {
		if !commandCodeSessionRe.MatchString(entry.Name()) {
			return nil
		}
		block.Total++
		started := firstRecordTimestamp(path)
		block.Observe(started, entryModTime(entry))
		return nil
	})
	if err != nil {
		return block, err
	}
	return block, nil
}

// Observe folds one session into the totals. started may be zero when the
// session's start time is unknown; modified is used for the newest activity.
func (b *SessionsBlock) Observe(started, modified time.Time) {
	if !started.IsZero() {
		utc := started.UTC()
		if b.Oldest == nil || utc.Before(*b.Oldest) {
			b.Oldest = &utc
		}
	}
	if !modified.IsZero() {
		utc := modified.UTC()
		if b.Newest == nil || utc.After(*b.Newest) {
			b.Newest = &utc
		}
	}
}

// walkFiles visits every regular file under root. A missing root is not an
// error: a provider that was never used simply has no sessions.
func walkFiles(root string, visit func(path string, entry fs.DirEntry) error) error {
	info, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if entry != nil && entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		return visit(path, entry)
	})
}

// firstRecordTimestamp reads the first JSON line of a transcript and returns
// its timestamp field, accepting both the "timestamp" and "createdAt" shapes.
func firstRecordTimestamp(path string) time.Time {
	file, err := os.Open(path)
	if err != nil {
		return time.Time{}
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	line, err := reader.ReadBytes('\n')
	if err != nil && len(line) == 0 {
		return time.Time{}
	}
	var record struct {
		Timestamp string `json:"timestamp"`
		CreatedAt string `json:"createdAt"`
	}
	if err := json.Unmarshal(line, &record); err != nil {
		return time.Time{}
	}
	if at := parseTimestamp(record.Timestamp); !at.IsZero() {
		return at
	}
	return parseTimestamp(record.CreatedAt)
}

func parseTimestamp(value string) time.Time {
	text := strings.TrimSpace(value)
	if text == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"} {
		if parsed, err := time.Parse(layout, text); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

// entryModTime returns a directory entry's modification time, or the zero time
// when it cannot be read.
func entryModTime(entry fs.DirEntry) time.Time {
	info, err := entry.Info()
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

func resolveHome(home, fallback string) string {
	if trimmed := strings.TrimSpace(home); trimmed != "" {
		return trimmed
	}
	return fallback
}
