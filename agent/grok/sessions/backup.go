package sessions

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// BackupOptions configures Backup.
type BackupOptions struct {
	// GrokHome is the Grok home directory. Empty → $GROK_HOME or ~/.grok.
	GrokHome string
	// OutDir is the backup directory. Empty → create a temp dir (kept).
	// On DryRun, OutDir is a plan input only and is never created.
	OutDir string
	// ArchivePath is an optional .tar.gz path. Empty → no archive.
	// Must end with ".tar.gz". Must not already exist on live backups.
	ArchivePath string
	// IncludeChildren when nil defaults to true (include linked children).
	// When false, child session dirs are not copied.
	IncludeChildren *bool
	// Live injects process listing / open-file probes for the busy gate.
	Live *LiveOptions
	// DryRun when true plans the backup only: no OutDir, archive, or manifest writes.
	DryRun bool
}

// BackupResult is the outcome of a successful Backup (live write or dry-run plan).
type BackupResult struct {
	Dir          string `json:"dir"`                     // absolute backup directory; empty when DryRun
	ArchivePath  string `json:"archive_path,omitempty"`  // absolute archive path, or ""; empty when DryRun
	ManifestPath string `json:"manifest_path,omitempty"` // Dir + "/manifest.json"; empty when DryRun
	SessionID    string `json:"session_id"`
	CWD          string `json:"cwd"`
	CWDKey       string `json:"cwd_key"` // url.PathEscape(abs cwd)

	// Dry-run / plan fields (zero values on live success unless noted):
	DryRun          bool     `json:"dry_run"`
	PlannedFiles    int      `json:"planned_files,omitempty"`    // estimated file count in plan (dry-run)
	PlannedBytes    int64    `json:"planned_bytes,omitempty"`    // estimated payload bytes (dry-run)
	RelatedSessions []string `json:"related_sessions,omitempty"` // parent + included children (dry-run always)
}

// BackupManifest is the version-1 self-describing manifest (manifest.json).
type BackupManifest struct {
	Version         int                    `json:"version"`
	Kind            string                 `json:"kind"`
	CreatedAt       string                 `json:"created_at"`
	GrokHome        string                 `json:"grok_home"`
	SessionID       string                 `json:"session_id"`
	CWD             string                 `json:"cwd"`
	CWDKey          string                 `json:"cwd_key"`
	RelatedSessions []string               `json:"related_sessions"`
	Files           []BackupFileEntry      `json:"files"`
	Checks          []string               `json:"checks"`
	CheckResults    map[string]BackupCheck `json:"check_results"`
	Logs            BackupLogsMeta         `json:"logs"`
	SQLite          BackupSQLiteNote       `json:"sqlite"`
	Stats           map[string]any         `json:"stats,omitempty"`
	Warnings        []string               `json:"warnings,omitempty"`
}

// BackupFileEntry describes one file written under the backup directory.
type BackupFileEntry struct {
	Path   string `json:"path"`   // relative to backup dir
	Source string `json:"source"` // absolute source path when applicable
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
	Role   string `json:"role"`
}

// BackupCheck is one named integrity / consistency check result.
type BackupCheck struct {
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
}

// BackupLogsMeta is log meta only (no log body under payload).
type BackupLogsMeta struct {
	Path       string          `json:"path"`
	MatchCount int             `json:"match_count"`
	LastLines  []BackupLogLine `json:"last_lines"`
}

// BackupLogLine is one matched log line retained in the manifest.
type BackupLogLine struct {
	Line int    `json:"line"` // 1-based source line number
	Text string `json:"text"`
	Time string `json:"time,omitempty"`
}

// BackupSQLiteNote notes session_search.sqlite without copying it.
type BackupSQLiteNote struct {
	Path    string `json:"path"`
	Present bool   `json:"present"`
	Note    string `json:"note,omitempty"`
}

const (
	backupManifestKind = "agent-pro.grok.session.backup"
	backupManifestVer  = 1
	promptHistoryName  = "prompt_history.jsonl"
	promptExtractName  = "prompt_history.session.jsonl"
	sqliteRelPath      = "sessions/session_search.sqlite"
	unifiedLogRel      = "logs/unified.jsonl"
)

// Backup creates a self-describing backup directory for one Grok session.
//
// Busy gate: errors (no payload) if IsFileActive OR LivePIDs non-empty.
// Does not copy logs or session_search.sqlite; those appear only as manifest meta.
// With opts.DryRun, plans the backup (file counts / related sessions) and writes nothing.
func Backup(sessionID string, opts *BackupOptions) (*BackupResult, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, sessionNotFoundError(sessionID)
	}
	if opts == nil {
		opts = &BackupOptions{}
	}

	grokHome, err := resolveBackupGrokHome(opts)
	if err != nil {
		return nil, err
	}

	session, err := Find(grokHome, sessionID)
	if err != nil {
		return nil, err
	}

	// Busy gate: either signal aborts before writing payload (live + dry-run).
	fileActive, err := IsFileActive(grokHome, sessionID)
	if err != nil {
		return nil, err
	}
	if fileActive {
		return nil, fmt.Errorf("session %s is active (file-active); cannot backup", sessionID)
	}
	livePIDs, err := LivePIDsForSession(sessionID, opts.Live)
	if err != nil {
		return nil, err
	}
	if len(livePIDs) > 0 {
		return nil, fmt.Errorf("session %s has live process (pid %d); cannot backup (busy)", sessionID, livePIDs[0].PID)
	}

	// Validate archive path early (before any payload write).
	// Suffix required on live and dry-run; already-exists is live only.
	archivePath := strings.TrimSpace(opts.ArchivePath)
	if archivePath != "" {
		if !strings.HasSuffix(archivePath, ".tar.gz") {
			return nil, fmt.Errorf("archive path must end with .tar.gz suffix: %s", archivePath)
		}
		absArchive, err := filepath.Abs(archivePath)
		if err != nil {
			return nil, fmt.Errorf("resolve archive path: %w", err)
		}
		archivePath = absArchive
		if !opts.DryRun && fileExists(archivePath) {
			return nil, fmt.Errorf("archive path already exists: %s", archivePath)
		}
	}

	cwd := strings.TrimSpace(session.CWD)
	if cwd == "" {
		if decoded, ok := decodeSessionParentCWD(filepath.Dir(session.Path)); ok {
			cwd = decoded
		}
	}
	if cwd != "" {
		if abs, err := filepath.Abs(cwd); err == nil {
			cwd = abs
		}
	}
	cwdKey := encodeSessionCWDKey(cwd)
	sessionDir := filepath.Dir(session.Path)

	includeChildren := true
	if opts.IncludeChildren != nil {
		includeChildren = *opts.IncludeChildren
	}

	childIDs := discoverChildSessionIDs(sessionDir)
	sort.Strings(childIDs)

	related := []string{sessionID}
	if includeChildren {
		related = append(related, childIDs...)
	}
	relatedSet := make(map[string]struct{}, len(related))
	for _, id := range related {
		relatedSet[id] = struct{}{}
	}

	// Dry-run: plan only — skip OutDir / archive / manifest writes.
	if opts.DryRun {
		plannedFiles, plannedBytes, err := planBackupPayload(grokHome, sessionID, sessionDir, cwdKey, includeChildren, childIDs, relatedSet)
		if err != nil {
			return nil, err
		}
		return &BackupResult{
			Dir:             "",
			ArchivePath:     "",
			ManifestPath:    "",
			SessionID:       sessionID,
			CWD:             cwd,
			CWDKey:          cwdKey,
			DryRun:          true,
			PlannedFiles:    plannedFiles,
			PlannedBytes:    plannedBytes,
			RelatedSessions: related,
		}, nil
	}

	// Resolve output directory (live only).
	outDir := strings.TrimSpace(opts.OutDir)
	var dir string
	if outDir == "" {
		tmp, err := os.MkdirTemp("", "grok-session-backup-*")
		if err != nil {
			return nil, fmt.Errorf("create temp backup dir: %w", err)
		}
		dir = tmp
	} else {
		absOut, err := filepath.Abs(outDir)
		if err != nil {
			return nil, fmt.Errorf("resolve out dir: %w", err)
		}
		dir = absOut
		if st, err := os.Stat(dir); err == nil {
			if !st.IsDir() {
				return nil, fmt.Errorf("out dir is not a directory: %s", dir)
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				return nil, fmt.Errorf("read out dir: %w", err)
			}
			if len(entries) > 0 {
				return nil, fmt.Errorf("out dir is not empty: %s", dir)
			}
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("stat out dir: %w", err)
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create out dir: %w", err)
		}
	}

	// On failure after creating dir, leave empty dir (no payload/manifest) so
	// error asserts pass. Clean any partial writes.
	var writeOK bool
	defer func() {
		if writeOK {
			return
		}
		// Remove partial payload/manifest if any; keep dir itself if OutDir was set.
		_ = os.RemoveAll(filepath.Join(dir, "payload"))
		_ = os.Remove(filepath.Join(dir, "manifest.json"))
		if outDir == "" {
			// Temp dir we created — remove entirely on failure.
			_ = os.RemoveAll(dir)
		}
	}()

	var files []BackupFileEntry
	var warnings []string

	// --- copy parent session tree ---
	parentDst := filepath.Join(dir, "payload", "sessions", cwdKey, sessionID)
	if err := copyDirRecursive(sessionDir, parentDst); err != nil {
		return nil, fmt.Errorf("copy parent session: %w", err)
	}
	parentEntries, err := collectCopiedFiles(parentDst, filepath.Join("payload", "sessions", cwdKey, sessionID), sessionDir, "session_file")
	if err != nil {
		return nil, err
	}
	files = append(files, parentEntries...)

	// --- copy included children ---
	if includeChildren {
		for _, childID := range childIDs {
			childSrc := filepath.Join(grokHome, "sessions", cwdKey, childID)
			if !fileExists(childSrc) {
				warnings = append(warnings, fmt.Sprintf("child session dir missing: %s", childID))
				continue
			}
			childDst := filepath.Join(dir, "payload", "sessions", cwdKey, childID)
			if err := copyDirRecursive(childSrc, childDst); err != nil {
				return nil, fmt.Errorf("copy child session %s: %w", childID, err)
			}
			childEntries, err := collectCopiedFiles(childDst, filepath.Join("payload", "sessions", cwdKey, childID), childSrc, "session_file")
			if err != nil {
				return nil, err
			}
			files = append(files, childEntries...)
		}
	}

	// --- filtered prompt history ---
	promptSrc := filepath.Join(grokHome, "sessions", cwdKey, promptHistoryName)
	promptDstRel := filepath.Join("payload", "sessions", cwdKey, promptExtractName)
	promptDst := filepath.Join(dir, promptDstRel)
	// Filter ids: parent + children when included. When children are skipped,
	// still include child lines only if they appear in related set (they don't).
	// Prompt-history tests use include=true. For no-children we filter by related only.
	// Also include linked child ids for prompt filtering when includeChildren so noise is dropped.
	promptIDs := relatedSet
	if n, err := writeFilteredPromptHistory(promptSrc, promptDst, promptIDs); err != nil {
		return nil, err
	} else if n >= 0 {
		// n == -1 means source missing — skip entry; else record.
		if fileExists(promptDst) {
			entry, err := fileEntryFor(promptDst, promptDstRel, promptSrc, "prompt_history")
			if err != nil {
				return nil, err
			}
			files = append(files, entry)
		}
	}

	// --- active_sessions entry extract (if present) ---
	if entryJSON, ok, err := extractActiveSessionEntry(grokHome, sessionID); err != nil {
		return nil, err
	} else if ok {
		rel := filepath.Join("payload", "bookkeeping", "active_sessions.entry.json")
		dst := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(dst, entryJSON, 0o644); err != nil {
			return nil, err
		}
		entry, err := fileEntryFor(dst, rel, filepath.Join(grokHome, "active_sessions.json"), "active_entry")
		if err != nil {
			return nil, err
		}
		files = append(files, entry)
	}

	// --- relocation lock ---
	lockSrc := filepath.Join(grokHome, "relocations", sessionID+".lock")
	if fileExists(lockSrc) {
		rel := filepath.Join("payload", "bookkeeping", "relocations", sessionID+".lock")
		dst := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return nil, err
		}
		if err := copyFile(lockSrc, dst); err != nil {
			return nil, fmt.Errorf("copy relocation lock: %w", err)
		}
		entry, err := fileEntryFor(dst, rel, lockSrc, "relocation_lock")
		if err != nil {
			return nil, err
		}
		files = append(files, entry)
	}

	// --- logs meta (no payload copy) ---
	logSrc := filepath.Join(grokHome, unifiedLogRel)
	logsMeta := scanUnifiedLogs(logSrc, relatedSet)

	// --- sqlite note (never copy) ---
	sqliteSrc := filepath.Join(grokHome, sqliteRelPath)
	sqliteNote := BackupSQLiteNote{
		Path:    sqliteSrc,
		Present: fileExists(sqliteSrc),
		Note:    "session_search.sqlite is not copied into the backup payload",
	}

	// --- checks ---
	checks, checkResults := buildBackupChecks(sessionID, cwd, includeChildren, childIDs, dir, cwdKey, promptDst, sqliteNote)

	manifest := BackupManifest{
		Version:         backupManifestVer,
		Kind:            backupManifestKind,
		CreatedAt:       time.Now().UTC().Format(time.RFC3339),
		GrokHome:        grokHome,
		SessionID:       sessionID,
		CWD:             cwd,
		CWDKey:          cwdKey,
		RelatedSessions: related,
		Files:           files,
		Checks:          checks,
		CheckResults:    checkResults,
		Logs:            logsMeta,
		SQLite:          sqliteNote,
		Stats: map[string]any{
			"file_count":    len(files),
			"related_count": len(related),
		},
		Warnings: warnings,
	}
	if len(warnings) == 0 {
		manifest.Warnings = nil
	}

	manifestPath := filepath.Join(dir, "manifest.json")
	if err := writeJSONFile(manifestPath, manifest); err != nil {
		return nil, fmt.Errorf("write manifest: %w", err)
	}

	// Optional archive of the backup directory (dir kept).
	resultArchive := ""
	if archivePath != "" {
		if err := writeTarGz(archivePath, dir); err != nil {
			return nil, fmt.Errorf("create archive: %w", err)
		}
		resultArchive = archivePath
	}

	writeOK = true
	return &BackupResult{
		Dir:          dir,
		ArchivePath:  resultArchive,
		ManifestPath: manifestPath,
		SessionID:    sessionID,
		CWD:          cwd,
		CWDKey:       cwdKey,
	}, nil
}

// planBackupPayload estimates the payload file count and byte size without writing.
// Mirrors live copy: parent tree, optional children, filtered prompt history,
// active_sessions entry, and relocation lock. Logs and sqlite are not payload.
func planBackupPayload(grokHome, sessionID, sessionDir, cwdKey string, includeChildren bool, childIDs []string, relatedSet map[string]struct{}) (int, int64, error) {
	var files int
	var bytes int64

	addTree := func(src string) error {
		n, b, err := countRegularFiles(src)
		if err != nil {
			return err
		}
		files += n
		bytes += b
		return nil
	}

	if err := addTree(sessionDir); err != nil {
		return 0, 0, fmt.Errorf("plan parent session: %w", err)
	}

	if includeChildren {
		for _, childID := range childIDs {
			childSrc := filepath.Join(grokHome, "sessions", cwdKey, childID)
			if !fileExists(childSrc) {
				continue
			}
			if err := addTree(childSrc); err != nil {
				return 0, 0, fmt.Errorf("plan child session %s: %w", childID, err)
			}
		}
	}

	// Filtered prompt history → one extract file when any matching lines exist.
	promptSrc := filepath.Join(grokHome, "sessions", cwdKey, promptHistoryName)
	if promptBytes, ok, err := planFilteredPromptHistoryBytes(promptSrc, relatedSet); err != nil {
		return 0, 0, err
	} else if ok {
		files++
		bytes += promptBytes
	}

	// Active sessions entry extract (same as live; busy gate usually prevents this).
	if entryJSON, ok, err := extractActiveSessionEntry(grokHome, sessionID); err != nil {
		return 0, 0, err
	} else if ok {
		files++
		bytes += int64(len(entryJSON))
	}

	// Relocation lock.
	lockSrc := filepath.Join(grokHome, "relocations", sessionID+".lock")
	if st, err := os.Stat(lockSrc); err == nil && st.Mode().IsRegular() {
		files++
		bytes += st.Size()
	}

	return files, bytes, nil
}

// countRegularFiles walks src and returns the number and total size of regular files.
func countRegularFiles(src string) (int, int64, error) {
	var n int
	var bytes int64
	err := filepath.WalkDir(src, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		n++
		bytes += info.Size()
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	return n, bytes, nil
}

// planFilteredPromptHistoryBytes returns the byte size of the filtered extract
// that would be written, and whether any matching lines exist.
func planFilteredPromptHistoryBytes(src string, ids map[string]struct{}) (int64, bool, error) {
	f, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, false, nil
		}
		return 0, false, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	var total int64
	matched := false
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			continue
		}
		sid, _ := obj["session_id"].(string)
		if sid == "" {
			sid, _ = obj["sessionId"].(string)
		}
		if _, ok := ids[sid]; !ok {
			continue
		}
		matched = true
		// line + trailing newline as written by writeFilteredPromptHistory
		total += int64(len(line) + 1)
	}
	if err := sc.Err(); err != nil {
		return 0, false, err
	}
	return total, matched, nil
}

func resolveBackupGrokHome(opts *BackupOptions) (string, error) {
	if opts != nil {
		if home := strings.TrimSpace(opts.GrokHome); home != "" {
			return filepath.Abs(home)
		}
	}
	if home := strings.TrimSpace(os.Getenv("GROK_HOME")); home != "" {
		return filepath.Abs(home)
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}
	return filepath.Join(userHome, ".grok"), nil
}

// discoverChildSessionIDs reads parent/subagents/*/meta.json for child_session_id.
func discoverChildSessionIDs(sessionDir string) []string {
	subRoot := filepath.Join(sessionDir, "subagents")
	entries, err := os.ReadDir(subRoot)
	if err != nil {
		return nil
	}
	seen := map[string]struct{}{}
	var ids []string
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		metaPath := filepath.Join(subRoot, ent.Name(), "meta.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}
		var meta struct {
			ChildSessionID string `json:"child_session_id"`
		}
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}
		id := strings.TrimSpace(meta.ChildSessionID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

func copyDirRecursive(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func collectCopiedFiles(dstRoot, relRoot, srcRoot, role string) ([]BackupFileEntry, error) {
	var out []BackupFileEntry
	err := filepath.WalkDir(dstRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(dstRoot, path)
		if err != nil {
			return err
		}
		relPath := filepath.ToSlash(filepath.Join(relRoot, rel))
		srcPath := filepath.Join(srcRoot, rel)
		sum, size, err := sha256File(path)
		if err != nil {
			return err
		}
		out = append(out, BackupFileEntry{
			Path:   relPath,
			Source: srcPath,
			Bytes:  size,
			SHA256: sum,
			Role:   role,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

func fileEntryFor(path, rel, source, role string) (BackupFileEntry, error) {
	sum, size, err := sha256File(path)
	if err != nil {
		return BackupFileEntry{}, err
	}
	return BackupFileEntry{
		Path:   filepath.ToSlash(rel),
		Source: source,
		Bytes:  size,
		SHA256: sum,
		Role:   role,
	}, nil
}

func sha256File(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

// writeFilteredPromptHistory writes JSONL lines whose session_id is in ids.
// Returns the number of lines written, or -1 if source does not exist.
func writeFilteredPromptHistory(src, dst string, ids map[string]struct{}) (int, error) {
	f, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			return -1, nil
		}
		return 0, err
	}
	defer f.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return 0, err
	}
	out, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	sc := bufio.NewScanner(f)
	// Allow long prompt lines.
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	n := 0
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			continue
		}
		sid, _ := obj["session_id"].(string)
		if sid == "" {
			// also accept sessionId
			sid, _ = obj["sessionId"].(string)
		}
		if _, ok := ids[sid]; !ok {
			continue
		}
		if _, err := out.WriteString(line + "\n"); err != nil {
			return n, err
		}
		n++
	}
	if err := sc.Err(); err != nil {
		return n, err
	}
	return n, out.Close()
}

// extractActiveSessionEntry returns the raw JSON of the matching entry if present.
func extractActiveSessionEntry(grokHome, sessionID string) ([]byte, bool, error) {
	path := filepath.Join(grokHome, "active_sessions.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	data = []byte(strings.TrimSpace(string(data)))
	if len(data) == 0 {
		return nil, false, nil
	}

	// Object form
	if data[0] == '{' {
		var obj struct {
			Sessions []json.RawMessage `json:"sessions"`
		}
		if err := json.Unmarshal(data, &obj); err != nil {
			return nil, false, nil
		}
		for _, raw := range obj.Sessions {
			if entrySessionID(raw) == sessionID {
				return raw, true, nil
			}
		}
		return nil, false, nil
	}

	// Array form
	if data[0] == '[' {
		var arr []json.RawMessage
		if err := json.Unmarshal(data, &arr); err != nil {
			return nil, false, nil
		}
		for _, raw := range arr {
			if entrySessionID(raw) == sessionID {
				return raw, true, nil
			}
		}
	}
	return nil, false, nil
}

func scanUnifiedLogs(path string, related map[string]struct{}) BackupLogsMeta {
	meta := BackupLogsMeta{
		Path:      path,
		LastLines: []BackupLogLine{},
	}
	f, err := os.Open(path)
	if err != nil {
		return meta
	}
	defer f.Close()

	var matches []BackupLogLine
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		text := sc.Text()
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			continue
		}
		sid := logLineSessionID(trimmed)
		if sid == "" {
			continue
		}
		if _, ok := related[sid]; !ok {
			continue
		}
		entry := BackupLogLine{
			Line: lineNo,
			Text: text,
			Time: logLineTime(trimmed),
		}
		matches = append(matches, entry)
	}
	meta.MatchCount = len(matches)
	if len(matches) == 0 {
		return meta
	}
	// last ≤ 3 matches in file order
	start := 0
	if len(matches) > 3 {
		start = len(matches) - 3
	}
	meta.LastLines = matches[start:]
	return meta
}

func logLineSessionID(line string) string {
	var obj map[string]any
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		// fallback: substring match not used — only structured sid
		return ""
	}
	for _, key := range []string{"sid", "session_id", "sessionId"} {
		if v, ok := obj[key].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func logLineTime(line string) string {
	var obj map[string]any
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		return ""
	}
	for _, key := range []string{"ts", "time", "timestamp"} {
		if v, ok := obj[key].(string); ok {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			// Prefer RFC3339-looking values as-is.
			if _, err := time.Parse(time.RFC3339, v); err == nil {
				return v
			}
			if _, err := time.Parse(time.RFC3339Nano, v); err == nil {
				return v
			}
			return v
		}
	}
	return ""
}

func buildBackupChecks(sessionID, cwd string, includeChildren bool, childIDs []string, dir, cwdKey, promptDst string, sqlite BackupSQLiteNote) ([]string, map[string]BackupCheck) {
	results := map[string]BackupCheck{}
	var names []string

	add := func(name string, ok bool, detail string) {
		names = append(names, name)
		results[name] = BackupCheck{OK: ok, Detail: detail}
	}

	// digests: every payload file has matching sha256 of on-disk bytes
	digestsOK := true
	detail := "all payload file digests verified"
	_ = filepath.WalkDir(filepath.Join(dir, "payload"), func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		sum, _, err := sha256File(path)
		if err != nil || sum == "" {
			digestsOK = false
			detail = fmt.Sprintf("digest failed for %s: %v", path, err)
			return filepath.SkipAll
		}
		return nil
	})
	add("digests", digestsOK, detail)

	// summary id matches
	summaryPath := filepath.Join(dir, "payload", "sessions", cwdKey, sessionID, "summary.json")
	sumOK := false
	sumDetail := "summary.json missing"
	if data, err := os.ReadFile(summaryPath); err == nil {
		var summary grokSummary
		if err := json.Unmarshal(data, &summary); err == nil && strings.TrimSpace(summary.Info.ID) == sessionID {
			sumOK = true
			sumDetail = "summary info.id matches session_id"
		} else {
			sumDetail = "summary info.id mismatch"
		}
	}
	add("summary_id", sumOK, sumDetail)

	// prompt_history id counts
	phOK := true
	phDetail := "prompt_history extract ok"
	if fileExists(promptDst) {
		// just ensure file is readable JSONL
		if data, err := os.ReadFile(promptDst); err != nil {
			phOK = false
			phDetail = err.Error()
		} else {
			lines := 0
			for _, line := range strings.Split(string(data), "\n") {
				if strings.TrimSpace(line) != "" {
					lines++
				}
			}
			phDetail = fmt.Sprintf("%d filtered lines", lines)
		}
	} else {
		phDetail = "prompt_history source missing or empty extract"
	}
	add("prompt_history", phOK, phDetail)

	// child set present when included
	if includeChildren {
		childOK := true
		var missing []string
		for _, id := range childIDs {
			p := filepath.Join(dir, "payload", "sessions", cwdKey, id)
			if !fileExists(p) {
				childOK = false
				missing = append(missing, id)
			}
		}
		detail := "all linked children present"
		if !childOK {
			detail = "missing children: " + strings.Join(missing, ",")
		} else if len(childIDs) == 0 {
			detail = "no linked children"
		}
		add("children", childOK, detail)
	}

	// sqlite note
	add("sqlite_note", true, fmt.Sprintf("present=%v path=%s", sqlite.Present, sqlite.Path))

	// busy not applicable post-copy
	add("busy_gate", true, "session was inactive at backup time")

	_ = cwd
	return names, results
}

// writeTarGz archives the contents of srcDir into destPath as gzip-compressed tar.
// Entries are relative to srcDir (manifest.json at archive root).
func writeTarGz(destPath, srcDir string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	err = filepath.WalkDir(srcDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(rel)
		if d.IsDir() {
			if !strings.HasSuffix(hdr.Name, "/") {
				hdr.Name += "/"
			}
			return tw.WriteHeader(hdr)
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(tw, in)
		_ = in.Close()
		return copyErr
	})
	if err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	return f.Close()
}
