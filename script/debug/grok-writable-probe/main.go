// grok-writable-probe captures grok TTY snapshots and compares writable heuristics.
//
// Starts (or reuses) a detached grok session via tty-watch, polls snapshot text,
// and records how each heuristic classifies writable state. Use the JSONL output
// to find false positives (e.g. "working tree" in scrollback).
//
// Run from repo root:
//
//	go run ./script/debug/grok-writable-probe
//	go run ./script/debug/grok-writable-probe -reuse -duration=30s
//	go run ./script/debug/grok-writable-probe -out=/tmp/grok-writable-capture
//	go run ./script/debug/grok-writable-probe -export-fixtures=pkgs/agenttty/testdata/grok-writable -from=/tmp/capture
//
// Requires tty-watch and grok on PATH (or set -tty-watch / -grok).
package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/agenttty"
	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
)

const defaultSessionID = "grok-debug"

type writableVerdict struct {
	Ready  bool   `json:"ready"`
	State  string `json:"state"`
	Reason string `json:"reason"`
}

type captureRecord struct {
	AtMS          int64           `json:"at_ms"`
	Phase         string          `json:"phase"`
	SnapshotLen   int             `json:"snapshot_len"`
	SnapshotHash  string          `json:"snapshot_hash"`
	SnapshotTail  string          `json:"snapshot_tail"`
	WorkingHits   []string        `json:"working_hits,omitempty"`
	ThinkingHits  []string        `json:"thinking_hits,omitempty"`
	HasPrompt     bool            `json:"has_prompt"`
	Current       writableVerdict `json:"current"`
	Tail12        writableVerdict `json:"tail12"`
	Tail24        writableVerdict `json:"tail24"`
	NoWorkingTree writableVerdict `json:"no_working_tree"`
	PromptRegion  writableVerdict `json:"prompt_region"`
	CodexStyle    writableVerdict `json:"codex_style"`
	CLIWritable   *writableVerdict `json:"cli_snapshot_match,omitempty"`
}

type fixtureExpectation struct {
	File   string   `json:"file"`
	Ready  bool     `json:"ready"`
	State  string   `json:"state"`
	Reason string   `json:"reason,omitempty"`
	Tags   []string `json:"tags"`
	Source string   `json:"source,omitempty"`
}

type phase struct {
	name    string
	message string // non-empty → tty-watch send before polling
	wait    time.Duration
}

func main() {
	os.Exit(run())
}

func run() int {
	var (
		sessionID      = flag.String("session-id", defaultSessionID, "tty-watch session id")
		outDir         = flag.String("out", "", "output directory (default: temp dir)")
		duration       = flag.Duration("duration", 45*time.Second, "total capture duration per phase after send")
		interval       = flag.Duration("interval", 250*time.Millisecond, "snapshot poll interval")
		reuse          = flag.Bool("reuse", false, "reuse existing session; do not start/kill grok-debug")
		killFirst      = flag.Bool("kill-first", true, "kill existing session before starting")
		ttyWatchBin    = flag.String("tty-watch", "", "path to tty-watch (default: PATH)")
		grokBin        = flag.String("grok", "", "path to grok binary (default: PATH or ~/.grok/bin/grok)")
		scenario       = flag.String("scenario", "follow-up", "baseline|follow-up|poll-only")
		exportFixtures = flag.String("export-fixtures", "", "export unique snapshots + expectations.jsonl to dir")
		fromDir        = flag.String("from", "", "read existing captures.jsonl + snapshots/ without probing")
	)
	flag.Parse()

	if *exportFixtures != "" {
		if *fromDir == "" {
			fmt.Fprintf(os.Stderr, "-export-fixtures requires -from=<capture-dir>\n")
			return 1
		}
		if err := exportFixturesFromCapture(*fromDir, *exportFixtures); err != nil {
			fmt.Fprintf(os.Stderr, "export fixtures: %v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "exported fixtures to %s\n", *exportFixtures)
		return 0
	}

	tw := *ttyWatchBin
	if tw == "" {
		var err error
		tw, err = exec.LookPath("tty-watch")
		if err != nil {
			fmt.Fprintf(os.Stderr, "tty-watch not found on PATH: %v\n", err)
			return 1
		}
	}
	grok := *grokBin
	if grok == "" {
		var err error
		grok, err = exec.LookPath("grok")
		if err != nil {
			home, _ := os.UserHomeDir()
			grok = filepath.Join(home, ".grok", "bin", "grok")
			if _, statErr := os.Stat(grok); statErr != nil {
				fmt.Fprintf(os.Stderr, "grok not found: %v\n", err)
				return 1
			}
		}
	}

	dir := *outDir
	if dir == "" {
		tmp, err := os.MkdirTemp("", "grok-writable-probe-*")
		if err != nil {
			fmt.Fprintf(os.Stderr, "mkdir temp: %v\n", err)
			return 1
		}
		dir = tmp
	} else if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir out: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "output: %s\n", dir)

	home, err := ttywatch.TTYWatchHome()
	if err != nil {
		fmt.Fprintf(os.Stderr, "tty-watch home: %v\n", err)
		return 1
	}

	if !*reuse {
		if *killFirst {
			_ = runCmd(tw, home, "kill", *sessionID)
			time.Sleep(300 * time.Millisecond)
		}
		stdout, stderr, code, err := runCmdOutput(tw, home, "run", "--detach", "--session-id="+*sessionID, grok, "--always-approve", "--permission-mode=bypassPermissions")
		if err != nil || code != 0 {
			fmt.Fprintf(os.Stderr, "start grok: exit=%d err=%v stderr=%s stdout=%s\n", code, err, stderr, stdout)
			return 1
		}
		fmt.Fprintf(os.Stderr, "started session %s\n", strings.TrimSpace(stdout))
	}

	entry, err := waitRegistry(home, *sessionID, 30*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wait registry: %v\n", err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "registry: listen=%s\n", entry.ListenAddr)

	phases := scenarioPhases(*scenario)
	capturesPath := filepath.Join(dir, "captures.jsonl")
	snapDir := filepath.Join(dir, "snapshots")
	if err := os.MkdirAll(snapDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir snapshots: %v\n", err)
		return 1
	}

	provider, _ := agenttty.Get("grok-tty")
	seenHashes := make(map[string]int)

	f, err := os.Create(capturesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create captures: %v\n", err)
		return 1
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(len(phases))*(*duration)+2*time.Minute)
	defer cancel()

	for _, ph := range phases {
		fmt.Fprintf(os.Stderr, "phase %q", ph.name)
		if ph.message != "" {
			fmt.Fprintf(os.Stderr, " send=%q", ph.message)
			if _, stderr, code, err := runCmdOutput(tw, home, "send", *sessionID, ph.message); err != nil || code != 0 {
				fmt.Fprintf(os.Stderr, "\n  send failed exit=%d err=%v stderr=%s\n", code, err, stderr)
			}
		}
		fmt.Fprintln(os.Stderr)

		phaseWait := ph.wait
		if phaseWait <= 0 {
			phaseWait = *duration
		}
		deadline := time.Now().Add(phaseWait)
		for time.Now().Before(deadline) {
			select {
			case <-ctx.Done():
				goto summarize
			default:
			}
			text, snapErr := ttywatch.SnapshotText(entry.ListenAddr, *sessionID)
			if snapErr != nil {
				fmt.Fprintf(os.Stderr, "snapshot error: %v\n", snapErr)
				time.Sleep(*interval)
				continue
			}
			rec := buildCapture(ph.name, text, provider.CheckWritable)
			if n := seenHashes[rec.SnapshotHash]; n < 3 {
				seenHashes[rec.SnapshotHash] = n + 1
				_ = os.WriteFile(filepath.Join(snapDir, fmt.Sprintf("%s_%s_%d.txt", ph.name, rec.SnapshotHash[:8], n)), []byte(text), 0644)
			}
			if cliText, cliErr := cliSnapshot(tw, home, *sessionID); cliErr == nil {
				match := cliText == text
				v := toVerdict(evalCurrent(cliText, provider.CheckWritable))
				if !match {
					v.Reason += " (cli text differs)"
				}
				rec.CLIWritable = &v
			}
			line, _ := json.Marshal(rec)
			_, _ = w.Write(line)
			_, _ = w.Write([]byte("\n"))
			_ = w.Flush()
			time.Sleep(*interval)
		}
	}

summarize:
	if err := writeSummary(capturesPath, filepath.Join(dir, "summary.txt")); err != nil {
		fmt.Fprintf(os.Stderr, "summary: %v\n", err)
	}
	fmt.Fprintf(os.Stderr, "done: %s\n", capturesPath)
	return 0
}

func scenarioPhases(name string) []phase {
	switch name {
	case "baseline":
		return []phase{{name: "idle", wait: 15 * time.Second}}
	case "poll-only":
		return []phase{{name: "poll", wait: 30 * time.Second}}
	default: // follow-up
		return []phase{
			{name: "boot", wait: 8 * time.Second},
			{name: "after_ls", message: "run ls and pwd"},
			{name: "after_recap", message: "what did I say?"},
			{name: "after_git", message: "run git status"},
			{name: "after_hello", message: "hello"},
		}
	}
}

func waitRegistry(home, sessionID string, timeout time.Duration) (*ttywatch.RegistryEntry, error) {
	cfg := ttywatch.DefaultRegistryConfig(home)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		entry, err := ttywatch.ReadRegistry(cfg, sessionID)
		if err == nil && entry != nil && entry.ListenAddr != "" && tcpReachable(entry.ListenAddr) {
			return entry, nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil, fmt.Errorf("session %s not reachable within %s", sessionID, timeout)
}

func tcpReachable(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func runCmd(bin, home string, args ...string) error {
	_, _, code, err := runCmdOutput(bin, home, args...)
	if err != nil {
		return err
	}
	if code != 0 {
		return fmt.Errorf("exit %d", code)
	}
	return nil
}

func runCmdOutput(bin, home string, args ...string) (stdout, stderr string, code int, err error) {
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), "TTY_WATCH_HOME="+home)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	if cmd.ProcessState != nil {
		code = cmd.ProcessState.ExitCode()
	}
	return outBuf.String(), errBuf.String(), code, err
}

func cliSnapshot(bin, home, sessionID string) (string, error) {
	stdout, stderr, code, err := runCmdOutput(bin, home, "snapshot", sessionID)
	if err != nil {
		return "", err
	}
	if code != 0 {
		return "", fmt.Errorf("snapshot exit %d: %s", code, stderr)
	}
	return strings.TrimRight(stdout, "\n"), nil
}

func buildCapture(phase, text string, check func([]byte) agenttty.WritableStatus) captureRecord {
	tail := text
	if len(tail) > 500 {
		tail = tail[len(tail)-500:]
	}
	hash := sha256.Sum256([]byte(text))
	return captureRecord{
		AtMS:         time.Now().UnixMilli(),
		Phase:        phase,
		SnapshotLen:  len(text),
		SnapshotHash: hex.EncodeToString(hash[:8]),
		SnapshotTail: tail,
		WorkingHits:  findSubstringContexts(text, "working"),
		ThinkingHits: findSubstringContexts(text, "thinking"),
		HasPrompt:    hasPromptMarker(text),
		Current:      toVerdict(evalCurrent(text, check)),
		Tail12:       toVerdict(evalOnTail(text, 12, checkGrokWritableTail)),
		Tail24:       toVerdict(evalOnTail(text, 24, checkGrokWritableTail)),
		NoWorkingTree: toVerdict(evalNoWorkingTree(text)),
		PromptRegion: toVerdict(evalPromptRegion(text)),
		CodexStyle:   toVerdict(evalCodexStyleBusy(text)),
	}
}

func evalCurrent(text string, check func([]byte) agenttty.WritableStatus) agenttty.WritableStatus {
	if check == nil {
		return agenttty.WritableStatus{Reason: "no check function"}
	}
	return check([]byte(text))
}

func toVerdict(st agenttty.WritableStatus) writableVerdict {
	return writableVerdict{Ready: st.Ready, State: st.State, Reason: st.Reason}
}

func hasPromptMarker(plain string) bool {
	return strings.Contains(plain, "\u203a") || strings.Contains(plain, "›") ||
		strings.Contains(plain, "❯") ||
		strings.Contains(plain, "Grok >") || strings.Contains(plain, "> ")
}

func tailLines(text string, n int) string {
	lines := strings.Split(text, "\n")
	if len(lines) <= n {
		return text
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}

// --- experimental heuristics (candidates for production fix) ---

func checkGrokWritableTail(scrollback []byte) agenttty.WritableStatus {
	plain := stripANSI(string(scrollback))
	if len(strings.TrimSpace(plain)) == 0 {
		return agenttty.WritableStatus{Reason: "no terminal output", State: "unknown"}
	}
	lower := strings.ToLower(plain)
	if strings.Contains(lower, "response:") || strings.Contains(lower, "submitted:") {
		return agenttty.WritableStatus{Ready: true, State: "idle"}
	}
	if strings.Contains(plain, "Grok \u203a") || strings.Contains(plain, "Grok ›") ||
		strings.Contains(plain, "Grok >") || hasPromptMarker(plain) {
		if strings.Contains(lower, "working") || strings.Contains(lower, "thinking") {
			return agenttty.WritableStatus{Reason: "agent still responding (tail)", State: "busy"}
		}
		return agenttty.WritableStatus{Ready: true, State: "idle"}
	}
	return agenttty.WritableStatus{Reason: "no prompt in tail", State: "unknown"}
}

func evalOnTail(full string, lines int, check func([]byte) agenttty.WritableStatus) agenttty.WritableStatus {
	return check([]byte(tailLines(full, lines)))
}

func evalNoWorkingTree(full string) agenttty.WritableStatus {
	scrubbed := strings.ReplaceAll(strings.ToLower(full), "working tree", "wt_scrubbed")
	return checkGrokWritableTail([]byte(scrubbed))
}

func evalPromptRegion(full string) agenttty.WritableStatus {
	region := promptRegion(full)
	return checkGrokWritableTail([]byte(region))
}

func promptRegion(full string) string {
	markers := []string{"Enter:send", "Shift+Tab:mode", "Ctrl+.:shortcuts", "╭---"}
	best := 0
	for _, m := range markers {
		if i := strings.LastIndex(full, m); i > best {
			best = i
		}
	}
	if best > 0 {
		start := best
		if start > 400 {
			start -= 400
		} else {
			start = 0
		}
		return full[start:]
	}
	return tailLines(full, 16)
}

func evalCodexStyleBusy(full string) agenttty.WritableStatus {
	plain := stripANSI(full)
	lower := strings.ToLower(plain)
	if !hasPromptMarker(plain) {
		return agenttty.WritableStatus{Reason: "no prompt", State: "unknown"}
	}
	if strings.Contains(lower, "•") && (strings.Contains(lower, "working") || strings.Contains(lower, "esc to interrupt")) {
		return agenttty.WritableStatus{Reason: "codex-style busy", State: "busy"}
	}
	if strings.Contains(lower, "thinking") && strings.Contains(lower, "thought for") {
		return agenttty.WritableStatus{Reason: "thought in progress", State: "busy"}
	}
	return agenttty.WritableStatus{Ready: true, State: "idle"}
}

func stripANSI(s string) string {
	// minimal: agenttty.StripANSI is not exported from another package path easily in script;
	// duplicate lightweight strip for probe-only analysis.
	var out strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && !((s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z')) {
				i++
			}
			if i < len(s) {
				i++
			}
			continue
		}
		out.WriteByte(s[i])
		i++
	}
	return out.String()
}

func findSubstringContexts(text, needle string) []string {
	lower := strings.ToLower(text)
	needle = strings.ToLower(needle)
	var hits []string
	start := 0
	for {
		i := strings.Index(lower[start:], needle)
		if i < 0 {
			break
		}
		idx := start + i
		from := idx - 30
		if from < 0 {
			from = 0
		}
		to := idx + len(needle) + 40
		if to > len(text) {
			to = len(text)
		}
		hits = append(hits, strings.TrimSpace(text[from:to]))
		start = idx + len(needle)
		if len(hits) >= 8 {
			break
		}
	}
	return hits
}

func exportFixturesFromCapture(captureDir, exportDir string) error {
	capturesPath := filepath.Join(captureDir, "captures.jsonl")
	data, err := os.ReadFile(capturesPath)
	if err != nil {
		return fmt.Errorf("read captures: %w", err)
	}
	snapDir := filepath.Join(captureDir, "snapshots")
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return fmt.Errorf("mkdir export dir: %w", err)
	}

	provider, _ := agenttty.Get("grok-tty")
	seen := make(map[string]captureRecord)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var rec captureRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			return fmt.Errorf("parse captures.jsonl: %w", err)
		}
		hash8 := rec.SnapshotHash
		if len(hash8) > 8 {
			hash8 = hash8[:8]
		}
		if _, ok := seen[hash8]; ok {
			continue
		}
		seen[hash8] = rec
	}

	manifestPath := filepath.Join(exportDir, "expectations.jsonl")
	manifest, err := os.Create(manifestPath)
	if err != nil {
		return fmt.Errorf("create expectations.jsonl: %w", err)
	}
	defer manifest.Close()
	mw := bufio.NewWriter(manifest)
	defer mw.Flush()

	for hash8, rec := range seen {
		snapshotPath, err := findSnapshotFile(snapDir, rec.Phase, hash8)
		if err != nil {
			return err
		}
		text, err := os.ReadFile(snapshotPath)
		if err != nil {
			return fmt.Errorf("read snapshot %s: %w", snapshotPath, err)
		}
		st := provider.CheckWritable(text)
		name := exportFixtureName(rec, hash8)
		if err := os.WriteFile(filepath.Join(exportDir, name), text, 0644); err != nil {
			return fmt.Errorf("write fixture %s: %w", name, err)
		}
		exp := fixtureExpectation{
			File:   name,
			Ready:  st.Ready,
			State:  st.State,
			Reason: st.Reason,
			Tags:   exportFixtureTags(rec),
			Source: fmt.Sprintf("grok-writable-probe export from %s", captureDir),
		}
		line, err := json.Marshal(exp)
		if err != nil {
			return err
		}
		if _, err := mw.Write(line); err != nil {
			return err
		}
		if _, err := mw.Write([]byte("\n")); err != nil {
			return err
		}
	}
	return nil
}

func findSnapshotFile(snapDir, phase, hash8 string) (string, error) {
	pattern := filepath.Join(snapDir, fmt.Sprintf("%s_%s_*.txt", phase, hash8))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no snapshot for phase=%s hash=%s in %s", phase, hash8, snapDir)
	}
	return matches[0], nil
}

func exportFixtureDetail(rec captureRecord) string {
	if rec.SnapshotLen == 0 {
		return "empty"
	}
	if rec.HasPrompt {
		return "prompt"
	}
	return "no-prompt"
}

func exportFixtureName(rec captureRecord, hash8 string) string {
	return fmt.Sprintf("grok-%s-%s-%s-%s.txt", rec.Phase, rec.Current.State, exportFixtureDetail(rec), hash8)
}

func exportFixtureTags(rec captureRecord) []string {
	tags := []string{rec.Phase, rec.Current.State}
	if detail := exportFixtureDetail(rec); detail != "" {
		tags = append(tags, detail)
	}
	return tags
}

func writeSummary(capturesPath, outPath string) error {
	data, err := os.ReadFile(capturesPath)
	if err != nil {
		return err
	}
	type counts struct {
		total, currentReady, tail12Ready, noWTReady, promptReady, disagree int
	}
	byPhase := make(map[string]*counts)
	var records []captureRecord
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var rec captureRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		records = append(records, rec)
		c := byPhase[rec.Phase]
		if c == nil {
			c = &counts{}
			byPhase[rec.Phase] = c
		}
		c.total++
		if rec.Current.Ready {
			c.currentReady++
		}
		if rec.Tail12.Ready {
			c.tail12Ready++
		}
		if rec.NoWorkingTree.Ready {
			c.noWTReady++
		}
		if rec.PromptRegion.Ready {
			c.promptReady++
		}
		if rec.Current.Ready != rec.NoWorkingTree.Ready {
			c.disagree++
		}
	}

	var b strings.Builder
	b.WriteString("grok writable probe summary\n")
	b.WriteString("===========================\n\n")
	for phase, c := range byPhase {
		fmt.Fprintf(&b, "phase %q: samples=%d\n", phase, c.total)
		fmt.Fprintf(&b, "  current ready:      %d/%d\n", c.currentReady, c.total)
		fmt.Fprintf(&b, "  tail12 ready:       %d/%d\n", c.tail12Ready, c.total)
		fmt.Fprintf(&b, "  no_working_tree:    %d/%d\n", c.noWTReady, c.total)
		fmt.Fprintf(&b, "  prompt_region:      %d/%d\n", c.promptReady, c.total)
		fmt.Fprintf(&b, "  current vs noWT disagree: %d/%d\n\n", c.disagree, c.total)
	}
	b.WriteString("Heuristic notes:\n")
	b.WriteString("- current: production checkGrokWritable on full snapshot\n")
	b.WriteString("- tail12/tail24: busy/working checks on last N lines only\n")
	b.WriteString("- no_working_tree: scrub 'working tree' before tail check\n")
	b.WriteString("- prompt_region: check only near input chrome (Enter:send / prompt box)\n")
	b.WriteString("- codex_style: require bullet+working like codex-tty\n\n")
	b.WriteString("Look for phases where current=busy but no_working_tree=idle — those are false positives.\n")

	// print worst examples
	for _, rec := range records {
		if rec.Current.State == "busy" && rec.NoWorkingTree.Ready {
			fmt.Fprintf(&b, "\nFALSE POSITIVE phase=%s hash=%s\n", rec.Phase, rec.SnapshotHash)
			fmt.Fprintf(&b, "  working_hits: %v\n", rec.WorkingHits)
			fmt.Fprintf(&b, "  tail: %q\n", rec.SnapshotTail)
		}
	}

	return os.WriteFile(outPath, []byte(b.String()), 0644)
}