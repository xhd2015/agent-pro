// Command migrate-sessions rewrites agent-run session storage from nested
// sessions/<runner>/<session_id>/ to flat sessions/<session_id>/.
//
// Intended build:
//
//	go build -o /tmp/migrate-sessions ./cmd/migrate-sessions
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: migrate-sessions [--home PATH] [--dry-run] [--backup-dir PATH]

Migrate nested agent-run session layout to flat sessions/<session_id>/.

Options:
  --home PATH         agent-run home (default: $AGENT_RUN_HOME or ~/.agent-run)
  --dry-run           print plan only; no moves or .layout write
  --backup-dir PATH   backup destination (default: <home>/backups/sessions-<timestamp>/)
  -h, --help          show help
`

type layoutFile struct {
	Version int `json:"version"`
}

type sessionMetaLite struct {
	Runner    string `json:"runner"`
	SessionID string `json:"session_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type nestedSession struct {
	Runner    string
	SessionID string
	SrcDir    string
	Meta      sessionMetaLite
	Updated   time.Time
}

type movePlan struct {
	Src        string
	Dst        string
	DstName    string // final directory name under sessions/
	Runner     string
	SessionID  string
	IsRename   bool // collision loser renamed to id__runner
	KeepNewer  bool // winner at bare id
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "migrate-sessions: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	var homeFlag string
	var dryRun bool
	var backupDir string
	_, err := flags.String("--home", &homeFlag).
		Bool("--dry-run", &dryRun).
		String("--backup-dir", &backupDir).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	home, err := resolveHome(homeFlag)
	if err != nil {
		return err
	}
	sessionsRoot := filepath.Join(home, "sessions")

	// Already flat?
	if isAlreadyFlat(sessionsRoot) {
		fmt.Printf("already flat: %s (layout v2 or flat session dirs); nothing to do\n", sessionsRoot)
		return nil
	}

	nested, err := scanNested(sessionsRoot)
	if err != nil {
		return err
	}
	if len(nested) == 0 {
		// No nested sessions; treat as no-op success (may be empty or flat without marker).
		if !dryRun {
			if err := writeLayoutV2(sessionsRoot); err != nil {
				return err
			}
		}
		fmt.Printf("no nested sessions under %s; wrote layout marker if needed\n", sessionsRoot)
		return nil
	}

	plans := planMoves(nested)
	if dryRun {
		fmt.Printf("dry-run: plan for %s (%d moves)\n", home, len(plans))
		for _, p := range plans {
			kind := "move"
			if p.IsRename {
				kind = "rename (collision)"
			}
			fmt.Printf("  %s: %s -> sessions/%s\n", kind, p.Src, p.DstName)
		}
		return nil
	}

	// Backup first
	if backupDir == "" {
		ts := time.Now().UTC().Format("20060102-150405")
		backupDir = filepath.Join(home, "backups", "sessions-"+ts)
	}
	if err := copyDir(sessionsRoot, backupDir); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}
	fmt.Printf("backup: %s\n", backupDir)

	var moved, renamed, skipped, errors int
	for _, p := range plans {
		if err := os.MkdirAll(filepath.Dir(p.Dst), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "error mkdir %s: %v\n", p.Dst, err)
			errors++
			continue
		}
		if _, err := os.Stat(p.Dst); err == nil {
			// Destination exists unexpectedly
			fmt.Fprintf(os.Stderr, "skip: destination exists %s\n", p.Dst)
			skipped++
			continue
		}
		if err := os.Rename(p.Src, p.Dst); err != nil {
			// Cross-device fallback: copy + remove
			if err2 := copyDir(p.Src, p.Dst); err2 != nil {
				fmt.Fprintf(os.Stderr, "error move %s -> %s: %v / %v\n", p.Src, p.Dst, err, err2)
				errors++
				continue
			}
			if err3 := os.RemoveAll(p.Src); err3 != nil {
				fmt.Fprintf(os.Stderr, "warning: moved via copy but failed to remove %s: %v\n", p.Src, err3)
			}
		}
		// Ensure meta.runner set from path runner if missing
		if err := ensureMetaRunner(p.Dst, p.Runner, p.SessionID); err != nil {
			fmt.Fprintf(os.Stderr, "warning: ensure meta.runner on %s: %v\n", p.Dst, err)
		}
		if p.IsRename {
			fmt.Printf("collision rename: %s -> sessions/%s\n", p.Src, p.DstName)
			renamed++
		} else {
			fmt.Printf("moved: %s -> sessions/%s\n", p.Src, p.DstName)
			moved++
		}
	}

	// Remove empty old runner directories
	_ = removeEmptyRunnerDirs(sessionsRoot)

	if err := writeLayoutV2(sessionsRoot); err != nil {
		return fmt.Errorf("write .layout: %w", err)
	}
	fmt.Printf("layout: sessions/.layout version=2\n")
	fmt.Printf("report: moved=%d renamed=%d skipped=%d errors=%d\n", moved, renamed, skipped, errors)
	if errors > 0 {
		return fmt.Errorf("%d errors during migration", errors)
	}
	return nil
}

func resolveHome(flag string) (string, error) {
	if strings.TrimSpace(flag) != "" {
		return filepath.Clean(flag), nil
	}
	if v := os.Getenv("AGENT_RUN_HOME"); v != "" {
		return filepath.Clean(v), nil
	}
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".agent-run"), nil
}

func isAlreadyFlat(sessionsRoot string) bool {
	data, err := os.ReadFile(filepath.Join(sessionsRoot, ".layout"))
	if err == nil {
		var lf layoutFile
		if json.Unmarshal(data, &lf) == nil && lf.Version >= 2 {
			return true
		}
	}
	entries, err := os.ReadDir(sessionsRoot)
	if err != nil {
		return false
	}
	hasSession := false
	for _, ent := range entries {
		if !ent.IsDir() || strings.HasPrefix(ent.Name(), ".") {
			continue
		}
		// Nested layout has runner dirs whose children are session dirs with meta.json.
		// Flat: top-level children have meta.json.
		meta := filepath.Join(sessionsRoot, ent.Name(), "meta.json")
		if _, err := os.Stat(meta); err == nil {
			hasSession = true
			continue
		}
		// If any top-level dir lacks meta.json, it may be a runner dir → not flat.
		// Check if it looks nested.
		sub, err := os.ReadDir(filepath.Join(sessionsRoot, ent.Name()))
		if err != nil {
			continue
		}
		for _, s := range sub {
			if s.IsDir() {
				if _, err := os.Stat(filepath.Join(sessionsRoot, ent.Name(), s.Name(), "meta.json")); err == nil {
					return false
				}
			}
		}
	}
	return hasSession
}

func scanNested(sessionsRoot string) ([]nestedSession, error) {
	entries, err := os.ReadDir(sessionsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []nestedSession
	for _, runnerEnt := range entries {
		if !runnerEnt.IsDir() || strings.HasPrefix(runnerEnt.Name(), ".") {
			continue
		}
		runner := runnerEnt.Name()
		// Skip if this looks like a flat session dir (has meta.json).
		if _, err := os.Stat(filepath.Join(sessionsRoot, runner, "meta.json")); err == nil {
			continue
		}
		sub, err := os.ReadDir(filepath.Join(sessionsRoot, runner))
		if err != nil {
			continue
		}
		for _, sessEnt := range sub {
			if !sessEnt.IsDir() {
				continue
			}
			sessionID := sessEnt.Name()
			src := filepath.Join(sessionsRoot, runner, sessionID)
			metaPath := filepath.Join(src, "meta.json")
			meta, updated := readMetaLite(metaPath, runner, sessionID)
			out = append(out, nestedSession{
				Runner:    runner,
				SessionID: sessionID,
				SrcDir:    src,
				Meta:      meta,
				Updated:   updated,
			})
		}
	}
	return out, nil
}

func readMetaLite(path, pathRunner, sessionID string) (sessionMetaLite, time.Time) {
	data, err := os.ReadFile(path)
	var meta sessionMetaLite
	if err == nil {
		_ = json.Unmarshal(data, &meta)
	}
	if strings.TrimSpace(meta.Runner) == "" {
		meta.Runner = pathRunner
	}
	if strings.TrimSpace(meta.SessionID) == "" {
		meta.SessionID = sessionID
	}
	t := parseTime(meta.UpdatedAt)
	if t.IsZero() {
		t = parseTime(meta.CreatedAt)
	}
	return meta, t
}

func parseTime(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	return time.Time{}
}

func planMoves(nested []nestedSession) []movePlan {
	// Group by session id
	byID := map[string][]nestedSession{}
	for _, n := range nested {
		byID[n.SessionID] = append(byID[n.SessionID], n)
	}
	var plans []movePlan
	for id, group := range byID {
		if len(group) == 1 {
			n := group[0]
			plans = append(plans, movePlan{
				Src:       n.SrcDir,
				Dst:       filepath.Join(filepath.Dir(filepath.Dir(n.SrcDir)), id),
				DstName:   id,
				Runner:    n.Runner,
				SessionID: id,
			})
			continue
		}
		// Collision: keep newer updated_at at bare id; rename losers to id__runner
		winner := 0
		for i := 1; i < len(group); i++ {
			if group[i].Updated.After(group[winner].Updated) {
				winner = i
			} else if group[i].Updated.Equal(group[winner].Updated) {
				// tie-break: prefer later created path lexicographically by runner for stability
				if group[i].Runner > group[winner].Runner {
					winner = i
				}
			}
		}
		for i, n := range group {
			if i == winner {
				plans = append(plans, movePlan{
					Src:       n.SrcDir,
					Dst:       filepath.Join(filepath.Dir(filepath.Dir(n.SrcDir)), id),
					DstName:   id,
					Runner:    n.Runner,
					SessionID: id,
					KeepNewer: true,
				})
				continue
			}
			name := id + "__" + n.Runner
			plans = append(plans, movePlan{
				Src:       n.SrcDir,
				Dst:       filepath.Join(filepath.Dir(filepath.Dir(n.SrcDir)), name),
				DstName:   name,
				Runner:    n.Runner,
				SessionID: id,
				IsRename:  true,
			})
		}
	}
	return plans
}

func ensureMetaRunner(sessionDir, runner, sessionID string) error {
	metaPath := filepath.Join(sessionDir, "meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return err
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	changed := false
	if r, _ := m["runner"].(string); strings.TrimSpace(r) == "" {
		m["runner"] = runner
		changed = true
	}
	if s, _ := m["session_id"].(string); strings.TrimSpace(s) == "" {
		m["session_id"] = sessionID
		changed = true
	}
	if !changed {
		return nil
	}
	out, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath, out, 0644)
}

func writeLayoutV2(sessionsRoot string) error {
	if err := os.MkdirAll(sessionsRoot, 0755); err != nil {
		return err
	}
	data, err := json.Marshal(layoutFile{Version: 2})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(sessionsRoot, ".layout"), append(data, '\n'), 0644)
}

func removeEmptyRunnerDirs(sessionsRoot string) error {
	entries, err := os.ReadDir(sessionsRoot)
	if err != nil {
		return err
	}
	for _, ent := range entries {
		if !ent.IsDir() || strings.HasPrefix(ent.Name(), ".") {
			continue
		}
		// Don't remove flat session dirs (have meta.json)
		dir := filepath.Join(sessionsRoot, ent.Name())
		if _, err := os.Stat(filepath.Join(dir, "meta.json")); err == nil {
			continue
		}
		// If empty or only leftover empty children, remove
		sub, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		if len(sub) == 0 {
			_ = os.Remove(dir)
		}
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}
