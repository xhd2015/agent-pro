// grok-session-cwd-rebind: experiment whether Grok session tool cwd can be
// rebound after the original workspace directory is deleted, by editing
// session meta files under a custom GROK_HOME.
//
// Modes:
//
//	(default) repro  — delete old workspace, resume with --cwd NEW, no meta
//	                   patch. Expect sticky/broken cwd (exit non-zero + REPRO:).
//	--expect=healthy — apply --patch set, then expect pwd tool cwd == NEW
//	                   (exit 0 + VERIFY:).
//
// Usage:
//
//	go run ./script/debug/grok-session-cwd-rebind
//	go run ./script/debug/grok-session-cwd-rebind --expect=healthy --patch=move_dir
//	go run ./script/debug/grok-session-cwd-rebind --keep-root /tmp/my-exp
//
// Minimal working patch (2026-07-13 experiment): **move_dir alone** rebinds tool
// cwd. Editing summary.json / prompt_context.json without moving the session
// directory under sessions/<url-encoded-cwd>/ is NOT enough.
//
// Prerequisites: `grok` on PATH; ~/.grok/auth.json (copied into custom GROK_HOME).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	expect := flag.String("expect", "repro", "repro | healthy")
	patchCSV := flag.String("patch", "", "comma list for healthy mode: summary,prompt_context,move_dir,chat_history (default healthy: move_dir)")
	timeout := flag.Duration("timeout", 3*time.Minute, "timeout per grok invocation")
	keepRoot := flag.String("keep-root", "", "if set, use this root dir instead of mktemp (not deleted)")
	grokBin := flag.String("grok", "", "path to grok (default: PATH)")
	flag.Parse()

	modeHealthy := strings.EqualFold(strings.TrimSpace(*expect), "healthy")
	if !modeHealthy && strings.TrimSpace(*expect) != "repro" {
		fmt.Fprintf(os.Stderr, "FAIL: --expect must be repro or healthy\n")
		os.Exit(2)
	}

	bin := strings.TrimSpace(*grokBin)
	if bin == "" {
		var err error
		bin, err = exec.LookPath("grok")
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: grok not on PATH: %v\n", err)
			os.Exit(2)
		}
	}

	authSrc := filepath.Join(os.Getenv("HOME"), ".grok", "auth.json")
	if _, err := os.Stat(authSrc); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: missing %s (copy/login required for custom GROK_HOME): %v\n", authSrc, err)
		os.Exit(2)
	}

	var root string
	var cleanup bool
	if strings.TrimSpace(*keepRoot) != "" {
		root = *keepRoot
		if err := os.MkdirAll(root, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: keep-root: %v\n", err)
			os.Exit(2)
		}
	} else {
		var err error
		root, err = os.MkdirTemp("", "grok-cwd-rebind-*")
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: temp root: %v\n", err)
			os.Exit(2)
		}
		cleanup = true
	}
	defer func() {
		if cleanup {
			_ = os.RemoveAll(root)
		} else {
			fmt.Printf("KEEP_ROOT: %s\n", root)
		}
	}()

	grokHome := filepath.Join(root, "grok-home")
	wsOld := filepath.Join(root, "ws-old")
	wsNew := filepath.Join(root, "ws-new")
	if err := os.MkdirAll(grokHome, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: grok-home: %v\n", err)
		os.Exit(2)
	}
	if err := os.MkdirAll(wsOld, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: ws-old: %v\n", err)
		os.Exit(2)
	}
	if err := copyFile(authSrc, filepath.Join(grokHome, "auth.json"), 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: seed auth: %v\n", err)
		os.Exit(2)
	}
	// Optional config.toml (non-fatal).
	_ = copyFile(filepath.Join(os.Getenv("HOME"), ".grok", "config.toml"), filepath.Join(grokHome, "config.toml"), 0o644)

	env := append(os.Environ(), "GROK_HOME="+grokHome)

	// --- Phase A: create session + pwd in ws-old ---
	fmt.Printf("PHASE_A: create session in %s\n", wsOld)
	outA, errA, err := runGrok(bin, env, *timeout, []string{
		"-p", pwdPrompt(),
		"--cwd", wsOld,
		"--always-approve",
		"--permission-mode=bypassPermissions",
		"--output-format", "plain",
	})
	writeEvidence(root, "out-a.txt", outA)
	writeEvidence(root, "err-a.txt", errA)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: phase A grok: %v\nstderr:\n%s\nstdout:\n%s\n", err, errA, outA)
		os.Exit(2)
	}
	oldCanon, _ := filepath.EvalSymlinks(wsOld)
	if oldCanon == "" {
		oldCanon = wsOld
	}
	// macOS often reports /private/tmp/...
	if !strings.Contains(outA, filepath.Base(wsOld)) && !strings.Contains(outA, oldCanon) {
		fmt.Fprintf(os.Stderr, "FAIL: phase A pwd did not mention ws-old\nstdout:\n%s\n", outA)
		os.Exit(2)
	}
	fmt.Printf("PHASE_A_OK: pwd mentions old workspace\n")

	sess, err := findNewestSession(grokHome)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: locate session: %v\n", err)
		os.Exit(2)
	}
	// recordedOldCWD is the session workspace at create time (before any patch).
	// Do not compare against sess.CWD after meta patches rewrite it.
	recordedOldCWD := sess.CWD
	fmt.Printf("SESSION_ID: %s\n", sess.ID)
	fmt.Printf("SESSION_CWD: %s\n", sess.CWD)
	fmt.Printf("SESSION_DIR: %s\n", sess.Dir)
	writeEvidence(root, "session_id.txt", sess.ID+"\n")
	writeEvidence(root, "session_cwd.txt", sess.CWD+"\n")

	// --- Phase B: delete old workspace, create new ---
	if err := os.RemoveAll(wsOld); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: remove ws-old: %v\n", err)
		os.Exit(2)
	}
	if err := os.MkdirAll(wsNew, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: ws-new: %v\n", err)
		os.Exit(2)
	}
	newCanon, _ := filepath.EvalSymlinks(wsNew)
	if newCanon == "" {
		newCanon = wsNew
	}
	// Prefer /private form if that is what grok uses for tmp.
	if strings.HasPrefix(recordedOldCWD, "/private/") && !strings.HasPrefix(newCanon, "/private/") {
		if abs, err := filepath.Abs(wsNew); err == nil {
			// macOS: /var and /tmp often live under /private
			if strings.HasPrefix(abs, "/var/") || strings.HasPrefix(abs, "/tmp/") {
				newCanon = "/private" + abs
			} else {
				newCanon = abs
			}
		}
	}

	patches := parsePatches(*patchCSV, modeHealthy)
	if modeHealthy {
		fmt.Printf("PHASE_PATCH: applying %v\n", patches)
		if err := applyPatches(sess, newCanon, patches); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: patch: %v\n", err)
			os.Exit(2)
		}
		// Reload session path after move_dir (meta may now say NEW).
		if sess2, err := findSessionByID(grokHome, sess.ID); err == nil {
			sess = sess2
			fmt.Printf("SESSION_DIR_AFTER_PATCH: %s\n", sess.Dir)
			fmt.Printf("SESSION_CWD_AFTER_PATCH: %s\n", sess.CWD)
		}
	} else {
		fmt.Printf("PHASE_PATCH: skipped (repro mode)\n")
	}

	// --- Phase C: resume with --cwd NEW ---
	fmt.Printf("PHASE_C: resume %s with --cwd %s\n", sess.ID, wsNew)
	outC, errC, err := runGrok(bin, env, *timeout, []string{
		"-p", pwdPrompt(),
		"--resume", sess.ID,
		"--cwd", wsNew,
		"--always-approve",
		"--permission-mode=bypassPermissions",
		"--output-format", "plain",
	})
	writeEvidence(root, "out-c.txt", outC)
	writeEvidence(root, "err-c.txt", errC)
	// Do not fail hard on non-zero: symptom analysis decides.

	combined := outC + "\n" + errC
	// Mentions of the ORIGINAL (pre-patch) workspace only.
	oldStill := pathMentioned(combined, recordedOldCWD) || pathMentioned(combined, oldCanon)
	newOK := pathMentioned(combined, newCanon) || pathMentioned(combined, wsNew)
	workspaceGone := strings.Contains(strings.ToLower(combined), "workspace path may be gone") ||
		strings.Contains(strings.ToLower(combined), "workspace directory is missing") ||
		strings.Contains(strings.ToLower(combined), "originally in")

	// Tool current_dir from latest updates under session dirs.
	toolDir := latestToolCurrentDir(grokHome, sess.ID)
	fmt.Printf("TOOL_CURRENT_DIR: %q\n", toolDir)
	fmt.Printf("RECORDED_OLD_CWD: %q\n", recordedOldCWD)
	fmt.Printf("OUT_MENTIONS_OLD: %v\n", oldStill)
	fmt.Printf("OUT_MENTIONS_NEW: %v\n", newOK)
	fmt.Printf("WORKSPACE_GONE_MSG: %v\n", workspaceGone)
	fmt.Printf("RUN_ERR: %v\n", err)

	// Healthy: shell tool cwd is NEW (not the deleted OLD). Chat history may still
	// mention OLD paths; that alone is not failure if tools bind to NEW.
	toolOnNew := toolDir != "" && (pathsEqual(toolDir, newCanon) || pathsEqual(toolDir, wsNew))
	toolOnOld := toolDir != "" && (pathsEqual(toolDir, recordedOldCWD) || pathsEqual(toolDir, oldCanon))
	last := lastNonEmptyLine(outC)
	lastOnNew := pathsEqual(last, newCanon) || pathsEqual(last, wsNew)
	lastOnOld := pathsEqual(last, recordedOldCWD) || pathsEqual(last, oldCanon)
	healthy := (toolOnNew || (toolDir == "" && lastOnNew)) && !toolOnOld && !lastOnOld

	// Symptom for repro: sticky/broken cwd after delete + resume without meta fix.
	symptom := false
	if toolOnOld {
		symptom = true
	}
	if workspaceGone {
		symptom = true
	}
	if toolDir == "" && lastOnOld {
		symptom = true
	}
	if strings.Contains(errC, "originally in") && strings.Contains(errC, filepath.Base(wsOld)) {
		symptom = true
	}
	// No tool dir and no clean new pwd → inconclusive failure for healthy, symptom for repro
	if !modeHealthy && !healthy && (oldStill || toolDir == "") {
		symptom = true
	}

	fmt.Printf("SYMPTOM: %v\n", symptom)
	fmt.Printf("HEALTHY: %v\n", healthy)
	fmt.Printf("ROOT: %s\n", root)

	if modeHealthy {
		if healthy {
			fmt.Printf("PASS: VERIFY pwd/tool cwd rebound to new workspace after patch=%v\n", patches)
			fmt.Printf("VERIFY: tool_current_dir=%q new=%q\n", toolDir, newCanon)
			cleanup = strings.TrimSpace(*keepRoot) == ""
			os.Exit(0)
		}
		fmt.Printf("FAIL: expected healthy rebound; toolDir=%q new=%q old=%q\n", toolDir, newCanon, recordedOldCWD)
		fmt.Printf("stdout:\n%s\nstderr:\n%s\n", trim(outC, 2000), trim(errC, 2000))
		cleanup = false // keep evidence
		os.Exit(1)
	}

	// repro mode
	if symptom {
		fmt.Printf("REPRO: sticky/broken session cwd after workspace delete + resume --cwd NEW (no meta patch)\n")
		fmt.Printf("REPRO: session_id=%s recorded_cwd=%s\n", sess.ID, recordedOldCWD)
		fmt.Printf("REPRO: tool_current_dir=%q workspace_gone_msg=%v out_mentions_old=%v\n", toolDir, workspaceGone, oldStill)
		if snip := firstLineContaining(combined, "originally in"); snip != "" {
			fmt.Printf("REPRO: %s\n", snip)
		}
		if snip := firstLineContaining(combined, "workspace"); snip != "" {
			low := strings.ToLower(snip)
			if strings.Contains(low, "gone") || strings.Contains(low, "missing") {
				fmt.Printf("REPRO: %s\n", snip)
			}
		}
		cleanup = false // keep evidence for loop log
		os.Exit(1)      // non-zero = REPRO PASS for bug-repro inspect
	}

	fmt.Printf("FAIL: expected sticky-cwd symptom but not observed (unexpected healthy or inconclusive)\n")
	fmt.Printf("stdout:\n%s\nstderr:\n%s\n", trim(outC, 2000), trim(errC, 2000))
	cleanup = false
	os.Exit(2)
}

func pwdPrompt() string {
	return "Use the shell tool to run exactly: pwd\nReply with a single line that is ONLY the absolute path from that command (no markdown)."
}

func runGrok(bin string, env []string, timeout time.Duration, args []string) (stdout, stderr string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = env
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

type sessionInfo struct {
	ID  string
	CWD string
	Dir string
}

func findNewestSession(grokHome string) (sessionInfo, error) {
	root := filepath.Join(grokHome, "sessions")
	var best sessionInfo
	var bestMod time.Time
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		if info.Name() != "summary.json" {
			return nil
		}
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		var sum struct {
			Info struct {
				ID  string `json:"id"`
				CWD string `json:"cwd"`
			} `json:"info"`
		}
		if json.Unmarshal(b, &sum) != nil || sum.Info.ID == "" {
			return nil
		}
		if info.ModTime().After(bestMod) {
			bestMod = info.ModTime()
			best = sessionInfo{ID: sum.Info.ID, CWD: sum.Info.CWD, Dir: filepath.Dir(path)}
		}
		return nil
	})
	if err != nil {
		return sessionInfo{}, err
	}
	if best.ID == "" {
		return sessionInfo{}, fmt.Errorf("no summary.json under %s", root)
	}
	return best, nil
}

func findSessionByID(grokHome, id string) (sessionInfo, error) {
	root := filepath.Join(grokHome, "sessions")
	var found sessionInfo
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() || info.Name() != "summary.json" {
			return nil
		}
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		var sum struct {
			Info struct {
				ID  string `json:"id"`
				CWD string `json:"cwd"`
			} `json:"info"`
		}
		if json.Unmarshal(b, &sum) != nil {
			return nil
		}
		if sum.Info.ID == id {
			found = sessionInfo{ID: sum.Info.ID, CWD: sum.Info.CWD, Dir: filepath.Dir(path)}
		}
		return nil
	})
	if found.ID == "" {
		return sessionInfo{}, fmt.Errorf("session %s not found", id)
	}
	return found, nil
}

func parsePatches(csv string, healthy bool) []string {
	csv = strings.TrimSpace(csv)
	if csv == "" {
		if healthy {
			// Minimal sufficient set: relocate session under encoded NEW cwd key.
			return []string{"move_dir"}
		}
		return nil
	}
	var out []string
	for _, p := range strings.Split(csv, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func applyPatches(sess sessionInfo, newCWD string, patches []string) error {
	// Order: file edits first, then move_dir last.
	var doMove bool
	for _, p := range patches {
		switch p {
		case "summary":
			if err := patchSummaryCWD(filepath.Join(sess.Dir, "summary.json"), newCWD); err != nil {
				return err
			}
		case "prompt_context":
			if err := patchPromptContextWD(filepath.Join(sess.Dir, "prompt_context.json"), newCWD); err != nil {
				return err
			}
		case "chat_history":
			if err := patchChatHistoryWorkspace(filepath.Join(sess.Dir, "chat_history.jsonl"), sess.CWD, newCWD); err != nil {
				return err
			}
		case "move_dir":
			doMove = true
		default:
			return fmt.Errorf("unknown patch %q", p)
		}
	}
	if doMove {
		return moveSessionDir(sess, newCWD)
	}
	return nil
}

func patchSummaryCWD(path, newCWD string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	info, _ := m["info"].(map[string]any)
	if info == nil {
		info = map[string]any{}
		m["info"] = info
	}
	info["cwd"] = newCWD
	if gr, ok := m["git_root_dir"].(string); ok && gr != "" {
		m["git_root_dir"] = newCWD + "/"
	}
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}

func patchPromptContextWD(path, newCWD string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	m["working_directory"] = newCWD
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}

func patchChatHistoryWorkspace(path, oldCWD, newCWD string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	// Best-effort string replace for workspace path mentions.
	s := string(b)
	s = strings.ReplaceAll(s, oldCWD, newCWD)
	// Also /tmp vs /private/tmp variants.
	if strings.HasPrefix(oldCWD, "/private") {
		s = strings.ReplaceAll(s, strings.TrimPrefix(oldCWD, "/private"), strings.TrimPrefix(newCWD, "/private"))
	}
	return os.WriteFile(path, []byte(s), 0o644)
}

func moveSessionDir(sess sessionInfo, newCWD string) error {
	// Parent of session dir is URL-encoded cwd key.
	parent := filepath.Dir(sess.Dir)
	sessionsRoot := filepath.Dir(parent)
	// Encode like Python urllib.quote(path, safe='') ≈ path escape with %2F
	key := pathEncode(newCWD)
	newParent := filepath.Join(sessionsRoot, key)
	if err := os.MkdirAll(newParent, 0o755); err != nil {
		return err
	}
	dest := filepath.Join(newParent, sess.ID)
	if dest == sess.Dir {
		return nil
	}
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("destination already exists: %s", dest)
	}
	if err := os.Rename(sess.Dir, dest); err != nil {
		// Cross-device fallback: copy not implemented; report.
		return fmt.Errorf("rename session dir: %w", err)
	}
	// Leave empty old parent if empty.
	_ = os.Remove(parent)
	return nil
}

func pathEncode(p string) string {
	// Match grok session layout: %2F for every /
	var b strings.Builder
	for _, r := range p {
		if r == '/' {
			b.WriteString("%2F")
		} else if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		} else {
			// percent-encode byte
			for _, by := range []byte(string(r)) {
				b.WriteString(fmt.Sprintf("%%%02X", by))
			}
		}
	}
	return b.String()
}

func latestToolCurrentDir(grokHome, sessionID string) string {
	var last string
	root := filepath.Join(grokHome, "sessions")
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}
		if info.Name() != "updates.jsonl" || !strings.Contains(path, sessionID) {
			return nil
		}
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		for _, line := range strings.Split(string(b), "\n") {
			if !strings.Contains(line, "current_dir") {
				continue
			}
			// cheap extract
			const key = `"current_dir":"`
			if i := strings.Index(line, key); i >= 0 {
				rest := line[i+len(key):]
				if j := strings.Index(rest, `"`); j >= 0 {
					last = rest[:j]
				}
			}
		}
		return nil
	})
	return last
}

func pathMentioned(text, p string) bool {
	if p == "" {
		return false
	}
	if strings.Contains(text, p) {
		return true
	}
	// /private/tmp vs /tmp
	alts := []string{p}
	if strings.HasPrefix(p, "/private") {
		alts = append(alts, strings.TrimPrefix(p, "/private"))
	} else if strings.HasPrefix(p, "/tmp/") {
		alts = append(alts, "/private"+p)
	}
	for _, a := range alts {
		if a != "" && strings.Contains(text, a) {
			return true
		}
	}
	return false
}

func pathsEqual(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	if a == b {
		return true
	}
	ra, _ := filepath.EvalSymlinks(a)
	rb, _ := filepath.EvalSymlinks(b)
	if ra != "" && rb != "" && ra == rb {
		return true
	}
	// /private prefix
	na := strings.TrimPrefix(a, "/private")
	nb := strings.TrimPrefix(b, "/private")
	return na == nb
}

func lastNonEmptyLine(s string) string {
	lines := strings.Split(s, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		t := strings.TrimSpace(lines[i])
		if t != "" {
			return t
		}
	}
	return ""
}

func firstLineContaining(s, sub string) string {
	for _, line := range strings.Split(s, "\n") {
		if strings.Contains(line, sub) {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

func trim(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func writeEvidence(root, name, body string) {
	_ = os.WriteFile(filepath.Join(root, name), []byte(body), 0o644)
}

func copyFile(src, dst string, mode os.FileMode) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, mode)
}
