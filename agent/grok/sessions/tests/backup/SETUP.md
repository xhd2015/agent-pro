# Scenario

**Feature**: backup a Grok session to a self-describing directory via fixtures

```
# fixture grok home: sessions, prompt_history, optional active/logs/sqlite/locks
test harness -> seedBackupWorld
  -> sessions.Backup(sessionID, &BackupOptions{GrokHome, OutDir, ArchivePath, IncludeChildren, Live, DryRun})
  -> live: BackupResult {Dir, ArchivePath, ManifestPath} + manifest.json
  -> dry-run: BackupResult {DryRun, PlannedFiles, PlannedBytes, RelatedSessions}; write nothing
```

## Preconditions

- Package exports `Backup`, `BackupOptions`, `BackupResult`, and manifest-related
  types as locked in root DSN. **Dry-run fields** (`DryRun`, `PlannedFiles`,
  `PlannedBytes`, `RelatedSessions` on result) are Classic TDD RED until implementer.
- Busy gate uses existing `IsFileActive` + `LivePIDsForSession` via
  `opts.Live` injectables (same on dry-run).
- Tests never read real `~/.grok` or shell out to live `ps`/`lsof`.
- Encoding for cwd keys: `url.PathEscape(filepath.Abs(cwd))` (`/` → `%2F`).

## Steps

1. Root `Setup` allocates `req.TempDir` and `req.GrokHome = {temp}/.grok`.
2. Grouping/leaf `Setup` seeds sessions, bookkeeping, logs, output paths; may set `DryRun`.
3. Root `Run` maps injectables into `BackupOptions` and calls `sessions.Backup`.
4. Leaf `Assert` checks result, filesystem payload, manifest fields, plan fields, or errors.

## Context

- Canonical parent id: `019f283a-bbbb-7bbb-bbbb-bbbbbbbbbb01`
- Canonical child id: `019f283a-cccc-7ccc-cccc-cccccccccc02`
- Noise prompt id: `019f283a-dddd-7ddd-dddd-dddddddddd03`
- Fixture workspace cwd under temp: `ws-project` (abs path stored on Request).
- Helpers seed realistic `summary.json`, `subagents/<id>/meta.json`,
  `prompt_history.jsonl`, `logs/unified.jsonl`, sqlite marker, relocation lock.

```go
import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	fixtureBackupParentID = "019f283a-bbbb-7bbb-bbbb-bbbbbbbbbb01"
	fixtureBackupChildID  = "019f283a-cccc-7ccc-cccc-cccccccccc02"
	fixtureBackupNoiseID  = "019f283a-dddd-7ddd-dddd-dddddddddd03"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.TempDir = t.TempDir()
	req.GrokHome = filepath.Join(req.TempDir, ".grok")
	if err := os.MkdirAll(filepath.Join(req.GrokHome, "sessions"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(req.GrokHome, "logs"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(req.GrokHome, "relocations"), 0o755); err != nil {
		return err
	}
	if req.OpenFiles == nil {
		req.OpenFiles = map[int][]string{}
	}
	// Default inactive + no live pids (safe for success leaves).
	writeActiveSessions(t, req.GrokHome /* none */)
	return nil
}

func encodeCWD(t *testing.T, cwd string) string {
	t.Helper()
	abs, err := filepath.Abs(cwd)
	if err != nil {
		t.Fatalf("abs cwd %q: %v", cwd, err)
	}
	return url.PathEscape(abs)
}

func absPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatalf("abs %q: %v", p, err)
	}
	return abs
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s: %v", path, err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir parent %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readFileString(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func readJSONMap(t *testing.T, path string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return m
}

// writeActiveSessions writes object-form active_sessions.json listing the given ids.
func writeActiveSessions(t *testing.T, grokHome string, sessionIDs ...string) {
	t.Helper()
	entries := make([]map[string]any, 0, len(sessionIDs))
	for _, id := range sessionIDs {
		entries = append(entries, map[string]any{
			"sessionId": id,
			"cwd":       "/tmp/backup-fixture",
			"openedAt":  "2026-07-01T12:00:00Z",
		})
	}
	writeJSON(t, filepath.Join(grokHome, "active_sessions.json"), map[string]any{
		"sessions": entries,
	})
}

// writeSessionDir seeds summary.json (+ optional marker file) under
// sessions/<encode(cwd)>/<id>/. Returns absolute session directory.
func writeSessionDir(t *testing.T, grokHome, sessionID, cwd, title, marker string) string {
	t.Helper()
	absCwd := absPath(t, cwd)
	dir := filepath.Join(grokHome, "sessions", url.PathEscape(absCwd), sessionID)
	mustMkdir(t, dir)
	summary := map[string]any{
		"info": map[string]any{
			"id":  sessionID,
			"cwd": absCwd,
		},
		"generated_title":   title,
		"created_at":        "2026-07-01T10:00:00.000Z",
		"updated_at":        "2026-07-01T11:00:00.000Z",
		"last_active_at":    "2026-07-01T11:00:00.000Z",
		"num_messages":      2,
		"num_chat_messages": 1,
	}
	writeJSON(t, filepath.Join(dir, "summary.json"), summary)
	if marker != "" {
		mustWriteFile(t, filepath.Join(dir, "marker.txt"), marker)
	}
	// Minimal session body files so recursive copy is meaningful.
	mustWriteFile(t, filepath.Join(dir, "updates.jsonl"), `{"type":"test","session":"`+sessionID+`"}`+"\n")
	return dir
}

// linkChildSubagent writes parent/subagents/<childID>/meta.json with child_session_id.
func linkChildSubagent(t *testing.T, parentDir, parentID, childID string) {
	t.Helper()
	metaDir := filepath.Join(parentDir, "subagents", childID)
	mustMkdir(t, metaDir)
	meta := map[string]any{
		"subagent_id":       childID,
		"parent_session_id": parentID,
		"child_session_id":  childID,
		"subagent_type":     "general-purpose",
		"description":       "backup fixture child",
		"status":            "completed",
	}
	writeJSON(t, filepath.Join(metaDir, "meta.json"), meta)
}

// writePromptHistory writes sessions/<cwdKey>/prompt_history.jsonl lines.
func writePromptHistory(t *testing.T, grokHome, cwd string, lines []map[string]any) string {
	t.Helper()
	absCwd := absPath(t, cwd)
	path := filepath.Join(grokHome, "sessions", url.PathEscape(absCwd), "prompt_history.jsonl")
	mustMkdir(t, filepath.Dir(path))
	var b strings.Builder
	for _, line := range lines {
		raw, err := json.Marshal(line)
		if err != nil {
			t.Fatalf("marshal prompt line: %v", err)
		}
		b.Write(raw)
		b.WriteByte('\n')
	}
	mustWriteFile(t, path, b.String())
	return path
}

func writeUnifiedLog(t *testing.T, grokHome string, lines []string) string {
	t.Helper()
	path := filepath.Join(grokHome, "logs", "unified.jsonl")
	mustWriteFile(t, path, strings.Join(lines, "\n")+"\n")
	return path
}

func writeSQLiteMarker(t *testing.T, grokHome, marker string) string {
	t.Helper()
	path := filepath.Join(grokHome, "sessions", "session_search.sqlite")
	mustWriteFile(t, path, marker)
	return path
}

func writeRelocationLock(t *testing.T, grokHome, sessionID string) string {
	t.Helper()
	path := filepath.Join(grokHome, "relocations", sessionID+".lock")
	mustWriteFile(t, path, "")
	return path
}

// seedStandardWorld creates parent+child under a workspace cwd with shared
// bookkeeping defaults used by most success leaves. Sets identity fields on req.
func seedStandardWorld(t *testing.T, req *Request) {
	t.Helper()
	ws := filepath.Join(req.TempDir, "ws-project")
	mustMkdir(t, ws)
	req.CWD = absPath(t, ws)
	req.CWDKey = encodeCWD(t, req.CWD)
	req.SessionID = fixtureBackupParentID
	req.ChildSessionID = fixtureBackupChildID
	req.PromptNoiseID = fixtureBackupNoiseID
	req.ParentMarker = "PARENT-MARKER-v1"
	req.ChildMarker = "CHILD-MARKER-v1"
	req.SQLiteMarker = "SQLITE-MARKER-DO-NOT-COPY-v1"

	parentDir := writeSessionDir(t, req.GrokHome, req.SessionID, req.CWD, "backup parent", req.ParentMarker)
	_ = writeSessionDir(t, req.GrokHome, req.ChildSessionID, req.CWD, "backup child", req.ChildMarker)
	linkChildSubagent(t, parentDir, req.SessionID, req.ChildSessionID)

	writePromptHistory(t, req.GrokHome, req.CWD, []map[string]any{
		{"timestamp": "2026-07-01T10:00:00Z", "session_id": req.SessionID, "prompt": "parent prompt one"},
		{"timestamp": "2026-07-01T10:01:00Z", "session_id": req.ChildSessionID, "prompt": "child prompt"},
		{"timestamp": "2026-07-01T10:02:00Z", "session_id": req.PromptNoiseID, "prompt": "noise other session"},
		{"timestamp": "2026-07-01T10:03:00Z", "session_id": req.SessionID, "prompt": "parent prompt two"},
	})

	req.SQLitePath = writeSQLiteMarker(t, req.GrokHome, req.SQLiteMarker)
	req.RelocationPath = writeRelocationLock(t, req.GrokHome, req.SessionID)

	// 5 matching log lines for parent+child (noise sid ignored); last 3 are the
	// trailing matches in file order.
	writeUnifiedLog(t, req.GrokHome, []string{
		`{"ts":"2026-07-01T12:00:01Z","sid":"` + req.SessionID + `","msg":"log-a"}`,
		`{"ts":"2026-07-01T12:00:02Z","sid":"` + req.PromptNoiseID + `","msg":"noise"}`,
		`{"ts":"2026-07-01T12:00:03Z","sid":"` + req.ChildSessionID + `","msg":"log-b"}`,
		`{"ts":"2026-07-01T12:00:04Z","sid":"` + req.SessionID + `","msg":"log-c"}`,
		`{"ts":"2026-07-01T12:00:05Z","sid":"` + req.SessionID + `","msg":"log-d"}`,
		`{"ts":"2026-07-01T12:00:06Z","sid":"` + req.ChildSessionID + `","msg":"log-e"}`,
	})
	req.LogMatchCount = 5

	// Ensure inactive (success default).
	writeActiveSessions(t, req.GrokHome /* none */)
}

func sessionDir(t *testing.T, grokHome, cwd, sessionID string) string {
	t.Helper()
	return filepath.Join(grokHome, "sessions", encodeCWD(t, cwd), sessionID)
}

func payloadSessionDir(backupDir, cwdKey, sessionID string) string {
	return filepath.Join(backupDir, "payload", "sessions", cwdKey, sessionID)
}

func assertNoHarnessErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Run harness error: %v", err)
	}
}

func assertNoError(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err != nil {
		t.Fatalf("unexpected error: %v", resp.Err)
	}
}

func assertError(t *testing.T, resp *Response) {
	t.Helper()
	if resp.Err == nil {
		t.Fatal("expected error, got nil")
	}
}

func assertErrorContains(t *testing.T, resp *Response, substrs ...string) {
	t.Helper()
	assertError(t, resp)
	msg := resp.Err.Error()
	for _, s := range substrs {
		if !strings.Contains(msg, s) {
			t.Fatalf("error %q missing %q", msg, s)
		}
	}
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()
	st, err := os.Stat(path)
	if err != nil || !st.IsDir() {
		t.Fatalf("expected directory %q: %v", path, err)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		t.Fatalf("expected file %q: %v", path, err)
	}
}

func assertPathMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected path missing: %q", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %q: %v", path, err)
	}
}

func assertEqualString(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %q, want %q", field, got, want)
	}
}

func assertEqualInt(t *testing.T, field string, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %d, want %d", field, got, want)
	}
}

func assertEqualBool(t *testing.T, field string, got, want bool) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %v, want %v", field, got, want)
	}
}

func assertContains(t *testing.T, got, substr string) {
	t.Helper()
	if !strings.Contains(got, substr) {
		t.Fatalf("missing %q in:\n%s", substr, got)
	}
}

func assertNotContains(t *testing.T, got, substr string) {
	t.Helper()
	if strings.Contains(got, substr) {
		t.Fatalf("unexpected %q in:\n%s", substr, got)
	}
}

// loadManifest reads backupDir/manifest.json as a generic map.
func loadManifest(t *testing.T, backupDir string) map[string]any {
	t.Helper()
	return readJSONMap(t, filepath.Join(backupDir, "manifest.json"))
}

func assertSuccessfulBackup(t *testing.T, req *Request, resp *Response) string {
	t.Helper()
	assertNoError(t, resp)
	if resp.Result == nil {
		t.Fatal("expected non-nil BackupResult")
	}
	dir := resp.Result.Dir
	if dir == "" {
		t.Fatal("Result.Dir empty")
	}
	assertDirExists(t, dir)
	assertFileExists(t, filepath.Join(dir, "manifest.json"))
	assertDirExists(t, filepath.Join(dir, "payload"))
	if resp.Result.ManifestPath != "" {
		assertEqualString(t, "ManifestPath", filepath.Clean(resp.Result.ManifestPath), filepath.Clean(filepath.Join(dir, "manifest.json")))
	}
	assertEqualString(t, "SessionID", resp.Result.SessionID, req.SessionID)
	return dir
}

func assertManifestCore(t *testing.T, man map[string]any, req *Request) {
	t.Helper()
	// version may be float64 from JSON numbers
	ver, ok := man["version"].(float64)
	if !ok || int(ver) != 1 {
		t.Fatalf("manifest.version = %v, want 1", man["version"])
	}
	assertEqualString(t, "kind", strField(man, "kind"), "agent-pro.grok.session.backup")
	assertEqualString(t, "session_id", strField(man, "session_id"), req.SessionID)
	assertEqualString(t, "cwd", strField(man, "cwd"), req.CWD)
	assertEqualString(t, "cwd_key", strField(man, "cwd_key"), req.CWDKey)
	if gh := strField(man, "grok_home"); gh != "" && filepath.Clean(gh) != filepath.Clean(req.GrokHome) {
		t.Fatalf("grok_home = %q, want %q", gh, req.GrokHome)
	}
}

func strField(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func assertNoPayloadUnder(t *testing.T, outDir string) {
	t.Helper()
	if outDir == "" {
		return
	}
	// OutDir may be missing, empty, or not contain payload/manifest from a partial write.
	payload := filepath.Join(outDir, "payload")
	manifest := filepath.Join(outDir, "manifest.json")
	if _, err := os.Stat(payload); err == nil {
		t.Fatalf("expected no payload dir after error: %s", payload)
	}
	if _, err := os.Stat(manifest); err == nil {
		t.Fatalf("expected no manifest.json after error: %s", manifest)
	}
}

func walkHasSuffix(t *testing.T, root, suffix string) bool {
	t.Helper()
	found := false
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, suffix) || d.Name() == suffix {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func grokOpenPath(sessionID string) string {
	return "/Users/fixture/.grok/sessions/%2Ftmp%2Fproj/" + sessionID + "/events.jsonl"
}

func assertFileEqualsMarker(t *testing.T, path, want string) {
	t.Helper()
	got := readFileString(t, path)
	if strings.TrimSpace(got) != strings.TrimSpace(want) {
		t.Fatalf("file %s = %q, want %q", path, got, want)
	}
}

func asStringSlice(t *testing.T, v any) []string {
	t.Helper()
	if v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		t.Fatalf("expected string array, got %T %v", v, v)
	}
	out := make([]string, 0, len(arr))
	for _, x := range arr {
		s, _ := x.(string)
		out = append(out, s)
	}
	return out
}

func sliceContains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

// assertDryRunSuccess checks a successful dry-run plan: no write paths, DryRun flag,
// PlannedFiles > 0, and identity fields.
func assertDryRunSuccess(t *testing.T, req *Request, resp *Response) {
	t.Helper()
	assertNoError(t, resp)
	if resp.Result == nil {
		t.Fatal("expected non-nil BackupResult on dry-run success")
	}
	r := resp.Result
	if !r.DryRun {
		t.Fatal("Result.DryRun = false, want true")
	}
	if r.Dir != "" {
		t.Fatalf("Result.Dir should be empty on dry-run, got %q", r.Dir)
	}
	if r.ArchivePath != "" {
		t.Fatalf("Result.ArchivePath should be empty on dry-run, got %q", r.ArchivePath)
	}
	if r.ManifestPath != "" {
		t.Fatalf("Result.ManifestPath should be empty on dry-run, got %q", r.ManifestPath)
	}
	if r.PlannedFiles <= 0 {
		t.Fatalf("PlannedFiles = %d, want > 0", r.PlannedFiles)
	}
	if r.PlannedBytes < 0 {
		t.Fatalf("PlannedBytes = %d, want >= 0", r.PlannedBytes)
	}
	assertEqualString(t, "SessionID", r.SessionID, req.SessionID)
	if req.OutDir != "" {
		assertPathMissing(t, req.OutDir)
	}
	if req.ArchivePath != "" {
		// Dry-run must not create the archive path (leaves use fresh paths).
		if _, err := os.Stat(req.ArchivePath); err == nil {
			t.Fatalf("ArchivePath must not be written on dry-run: %s", req.ArchivePath)
		}
	}
	assertNoPayloadUnder(t, req.OutDir)
}
```
