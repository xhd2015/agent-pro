package agentruncli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
	"github.com/xhd2015/less-gen/flags"
)

const ptyHelp = `
Usage: agent-run pty <subcommand> [ARGS]

Subcommands:
  stats         print PTY limit / masters / free estimate and agent-run __serve summary
  kill-orphans  list or terminate orphan agent-run __serve processes

Options:
  -h, --help   show help

Run agent-run pty <subcommand> --help for subcommand-specific options.
`

const ptyStatsHelp = `
Usage: agent-run pty stats

Print a human-readable summary of PTY resources and agent-run __serve processes.

Best-effort probes (sysctl, lsof, ps) may be partial; partial stats still exit 0.

Options:
  -h, --help   show help
`

const ptyKillOrphansHelp = `
Usage: agent-run pty kill-orphans [OPTIONS]

List or terminate orphan agent-run __serve* processes.

Default selection: only serves with PPID == 1 (true orphans).
Use --kind to match by path/argv markers (any PPID), or --all for every serve.

Options:
  --dry-run                 list matching serves without killing
  --exe PATH                only match serves whose executable equals PATH
  --all                     match all agent-run __serve* processes (wins over --kind)
  --kind=test-generated     match serves whose argv/exe contains TestGenerated
  --kind=workdir-at-tmp     match serves whose argv/exe contains /var/folders/ and /T/
  -h, --help                show help

Multiple --kind values are OR-ed. --exe is always ANDed when set.
`

type ptyServeInfo struct {
	PID       int
	PPID      int
	Etime     string
	SessionID string
	Command   string
	Exe       string
	Category  string
}

func runPty(args []string) error {
	if len(args) == 0 {
		fmt.Print(strings.TrimPrefix(ptyHelp, "\n"))
		return nil
	}
	cmd := args[0]
	sub := args[1:]
	switch cmd {
	case "-h", "--help":
		fmt.Print(strings.TrimPrefix(ptyHelp, "\n"))
		return nil
	case "stats":
		return runPtyStats(sub)
	case "kill-orphans":
		return runPtyKillOrphans(sub)
	default:
		return fmt.Errorf("unknown pty subcommand: %s", cmd)
	}
}

func runPtyStats(args []string) error {
	_, err := flags.Help("-h,--help", ptyStatsHelp).Parse(args)
	if err != nil {
		return err
	}

	limit, limitOK := probePTMXMax()
	masters, mastersOK := probeUniquePTMXMasters()
	serves, listErr := listAgentRunServes("")
	if listErr != nil {
		return fmt.Errorf("pty stats: cannot list processes: %w", listErr)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "PTY stats\n")
	if limitOK {
		fmt.Fprintf(&b, "  ptmx limit (kern.tty.ptmx_max): %d\n", limit)
	} else {
		fmt.Fprintf(&b, "  ptmx limit: unknown\n")
	}
	if mastersOK {
		fmt.Fprintf(&b, "  unique /dev/ptmx masters: %d\n", masters)
		if limitOK {
			free := limit - masters
			if free < 0 {
				free = 0
			}
			fmt.Fprintf(&b, "  free estimate: %d\n", free)
		}
	} else {
		fmt.Fprintf(&b, "  unique masters: unknown\n")
	}

	cats := map[string]int{"test": 0, "brainstorm": 0, "seatalk": 0, "other": 0}
	for _, s := range serves {
		cats[s.Category]++
	}
	fmt.Fprintf(&b, "\nagent-run __serve processes: %d\n", len(serves))
	fmt.Fprintf(&b, "  categories: test=%d brainstorm=%d seatalk=%d other=%d\n",
		cats["test"], cats["brainstorm"], cats["seatalk"], cats["other"])
	if len(serves) > 0 {
		fmt.Fprintf(&b, "  holders (agent-run serves):\n")
		// Cap detail lines to keep output readable on large hosts.
		maxShow := 20
		for i, s := range serves {
			if i >= maxShow {
				fmt.Fprintf(&b, "  … and %d more\n", len(serves)-maxShow)
				break
			}
			fmt.Fprintf(&b, "  pid=%d session=%s category=%s etime=%s\n",
				s.PID, emptyDash(s.SessionID), s.Category, emptyDash(s.Etime))
		}
	}
	fmt.Print(b.String())
	return nil
}

func runPtyKillOrphans(args []string) error {
	var dryRun bool
	var all bool
	var exeFilter string
	var kinds []string
	remaining, err := flags.Bool("--dry-run", &dryRun).
		Bool("--all", &all).
		String("--exe", &exeFilter).
		StringSlice("--kind", &kinds).
		Help("-h,--help", ptyKillOrphansHelp).
		Parse(args)
	if err != nil {
		return err
	}
	if len(remaining) > 0 {
		return fmt.Errorf("pty kill-orphans: unexpected arguments: %s", strings.Join(remaining, " "))
	}

	// Validate --kind values up front (even when --all wins over kind matching).
	for _, k := range kinds {
		k = strings.TrimSpace(k)
		if k == "" {
			return fmt.Errorf("pty kill-orphans: empty --kind value")
		}
		if !isKnownKillOrphanKind(k) {
			return fmt.Errorf("pty kill-orphans: unknown kind %q (valid: test-generated, workdir-at-tmp)", k)
		}
	}

	serves, listErr := listAgentRunServes(strings.TrimSpace(exeFilter))
	if listErr != nil {
		return fmt.Errorf("pty kill-orphans: cannot list processes: %w", listErr)
	}

	self := os.Getpid()
	// Selection: --all → keep all; else if any --kind → OR kind predicates;
	// else default PPID == 1. --exe already applied in listAgentRunServes.
	// Never target self.
	filtered := serves[:0]
	for _, s := range serves {
		if s.PID == self {
			continue
		}
		if !serveMatchesKillOrphanFilter(s, all, kinds) {
			continue
		}
		filtered = append(filtered, s)
	}
	serves = filtered

	if len(serves) == 0 {
		fmt.Println("no matching serves (no orphans)")
		return nil
	}

	filterHint := killOrphanFilterHint(all, kinds)
	if dryRun {
		fmt.Fprintf(os.Stdout, "dry-run: would kill %d matching serve(s) (%s)\n", len(serves), filterHint)
		for _, s := range serves {
			fmt.Fprintf(os.Stdout, "  pid=%d session=%s cmd=%s\n",
				s.PID, emptyDash(s.SessionID), truncateCmd(s.Command, 120))
		}
		return nil
	}

	var killed []int
	var failed []int
	for _, s := range serves {
		if err := terminateServePID(s.PID); err != nil {
			failed = append(failed, s.PID)
			continue
		}
		killed = append(killed, s.PID)
	}
	fmt.Fprintf(os.Stdout, "killed %d serve process(es)", len(killed))
	if len(killed) > 0 {
		fmt.Fprintf(os.Stdout, ": %s", joinInts(killed))
	}
	fmt.Fprintln(os.Stdout)
	if len(failed) > 0 {
		return fmt.Errorf("failed to kill pid(s): %s", joinInts(failed))
	}
	return nil
}

func isKnownKillOrphanKind(kind string) bool {
	switch kind {
	case "test-generated", "workdir-at-tmp":
		return true
	default:
		return false
	}
}

// serveMatchesKillOrphanFilter applies selection after list+exe:
// --all wins; else any --kind (OR); else PPID == 1 only.
func serveMatchesKillOrphanFilter(s ptyServeInfo, all bool, kinds []string) bool {
	if all {
		return true
	}
	if len(kinds) > 0 {
		for _, k := range kinds {
			if serveMatchesKind(s, strings.TrimSpace(k)) {
				return true
			}
		}
		return false
	}
	return s.PPID == 1
}

func serveMatchesKind(s ptyServeInfo, kind string) bool {
	haystack := s.Command
	if s.Exe != "" && !strings.Contains(haystack, s.Exe) {
		haystack = s.Exe + " " + haystack
	}
	switch kind {
	case "test-generated":
		return strings.Contains(haystack, "TestGenerated")
	case "workdir-at-tmp":
		return strings.Contains(haystack, "/var/folders/") && strings.Contains(haystack, "/T/")
	default:
		return false
	}
}

func killOrphanFilterHint(all bool, kinds []string) string {
	if all {
		return "all"
	}
	if len(kinds) > 0 {
		return "kind:" + strings.Join(kinds, ",")
	}
	return "ppid1"
}

func emptyDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func truncateCmd(s string, n int) string {
	s = strings.TrimSpace(s)
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func joinInts(xs []int) string {
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = strconv.Itoa(x)
	}
	return strings.Join(parts, ", ")
}

func terminateServePID(pid int) error {
	if pid <= 0 || pid == os.Getpid() {
		return nil
	}
	if !processRunning(pid) {
		return nil
	}
	_ = syscall.Kill(pid, syscall.SIGTERM)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && processRunning(pid) {
		time.Sleep(50 * time.Millisecond)
	}
	if processRunning(pid) {
		_ = syscall.Kill(pid, syscall.SIGKILL)
		// Best-effort process group (Setsid serve re-exec).
		_ = syscall.Kill(-pid, syscall.SIGKILL)
		time.Sleep(100 * time.Millisecond)
	}
	if processRunning(pid) {
		return fmt.Errorf("pid %d still running", pid)
	}
	return nil
}

// processRunning reports whether pid is a live (non-zombie) process.
// Signal 0 succeeds for zombies; those no longer hold PTYs and count as gone.
func processRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	if !ttywatch.ProcessAlive(pid) {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	out, err := exec.CommandContext(ctx, "ps", "-p", strconv.Itoa(pid), "-o", "state=").Output()
	if err != nil {
		// ps failure with signal-0 success: fall back to "alive".
		return true
	}
	state := strings.TrimSpace(string(out))
	if state == "" || strings.HasPrefix(state, "Z") {
		return false
	}
	return true
}

func probePTMXMax() (int, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	// macOS; other platforms may lack this key.
	cmd := exec.CommandContext(ctx, "sysctl", "-n", "kern.tty.ptmx_max")
	out, err := cmd.Output()
	if err != nil {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

func probeUniquePTMXMasters() (int, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	// Count unique PIDs holding /dev/ptmx (best-effort; may need root for some systems).
	cmd := exec.CommandContext(ctx, "lsof", "-F", "p", "--", "/dev/ptmx")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil
	_ = cmd.Run()
	if stdout.Len() == 0 {
		return 0, false
	}
	seen := make(map[int]struct{})
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "p") {
			continue
		}
		pid, err := strconv.Atoi(line[1:])
		if err != nil || pid <= 0 {
			continue
		}
		seen[pid] = struct{}{}
	}
	if len(seen) == 0 {
		return 0, false
	}
	return len(seen), true
}

// listAgentRunServes returns live processes whose argv indicates an agent-run
// __serve re-exec. When exeFilter is non-empty, only serves whose executable
// path matches that path (after clean/resolve) are returned.
func listAgentRunServes(exeFilter string) ([]ptyServeInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Wide args so serve tokens and session ids are visible. Include state to skip zombies.
	cmd := exec.CommandContext(ctx, "ps", "-axww", "-o", "pid=,ppid=,etime=,state=,command=")
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("ps timed out")
		}
		return nil, err
	}

	var filterResolved string
	if exeFilter != "" {
		filterResolved = resolveExePath(exeFilter)
	}

	var serves []ptyServeInfo
	for _, line := range strings.Split(string(out), "\n") {
		info, ok := parsePSServeLine(line)
		if !ok {
			continue
		}
		if info.Command == "" || strings.HasPrefix(info.Command, "<defunct>") {
			continue
		}
		if !isAgentRunServeCommand(info.Command) {
			continue
		}
		info.Exe = firstArg(info.Command)
		info.SessionID = parseServeSessionID(info.Command)
		info.Category = categorizeServe(info.Command)
		if filterResolved != "" {
			if !exePathsMatch(info.Exe, filterResolved) {
				continue
			}
		}
		serves = append(serves, info)
	}
	sort.Slice(serves, func(i, j int) bool { return serves[i].PID < serves[j].PID })
	return serves, nil
}

func resolveExePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	if abs, err := filepath.Abs(p); err == nil {
		p = abs
	}
	p = filepath.Clean(p)
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved
	}
	return p
}

func exePathsMatch(got, want string) bool {
	got = strings.TrimSpace(got)
	want = strings.TrimSpace(want)
	if got == "" || want == "" {
		return false
	}
	// Direct string compare (common when ps shows full launch path).
	if got == want {
		return true
	}
	rg := resolveExePath(got)
	rw := resolveExePath(want)
	if rg != "" && rw != "" && rg == rw {
		return true
	}
	// Basename fallback only when both look like absolute agent-run binaries
	// is too loose; prefer exact path. Also accept when command starts with want+space.
	if strings.HasPrefix(got, want) {
		rest := got[len(want):]
		if rest == "" || rest[0] == ' ' || rest[0] == '\t' {
			return true
		}
	}
	return false
}

func firstArg(command string) string {
	command = strings.TrimSpace(command)
	if command == "" {
		return ""
	}
	// Split on first space/tab; paths with spaces are rare for agent-run.
	for i := 0; i < len(command); i++ {
		if command[i] == ' ' || command[i] == '\t' {
			return command[:i]
		}
	}
	return command
}

func isAgentRunServeCommand(command string) bool {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return false
	}
	// Production re-exec: first arg after binary is IsServeSubcommand token.
	// fields[0] is exe; fields[1] is typically the serve token.
	for _, f := range fields {
		if ttywatch.IsServeSubcommand(f) {
			return true
		}
	}
	// Fallback: command line contains agent-run ... __serve...
	joined := strings.Join(fields, " ")
	if strings.Contains(joined, "__serve_") && strings.Contains(joined, "agent-run") {
		return true
	}
	return false
}

func parseServeSessionID(command string) string {
	fields := strings.Fields(command)
	// agent-run __serve_…__ <session-id> <cmd…>
	for i, f := range fields {
		if ttywatch.IsServeSubcommand(f) && i+1 < len(fields) {
			return fields[i+1]
		}
	}
	return ""
}

func categorizeServe(command string) string {
	lower := strings.ToLower(command)
	switch {
	case strings.Contains(lower, "brainstorm"):
		return "brainstorm"
	case strings.Contains(lower, "seatalk"):
		return "seatalk"
	case strings.Contains(lower, "testgenerated") ||
		strings.Contains(lower, "doctest") ||
		strings.Contains(lower, "pty-orphan") ||
		strings.Contains(lower, "/t/") ||
		strings.Contains(lower, "/tmp/") ||
		strings.Contains(lower, "go-build") ||
		strings.Contains(lower, "agent-run-pty-doctest"):
		return "test"
	default:
		return "other"
	}
}

// parsePSServeLine parses `ps -o pid=,ppid=,etime=,state=,command=` (state optional).
// etime can contain colons (e.g. 01:23 or 1-02:03:04) but no spaces when
// using etime= (no header padding between fields after trim).
func parsePSServeLine(line string) (ptyServeInfo, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return ptyServeInfo{}, false
	}
	// pid
	i := 0
	for i < len(line) && line[i] >= '0' && line[i] <= '9' {
		i++
	}
	if i == 0 {
		return ptyServeInfo{}, false
	}
	pid, err := strconv.Atoi(line[:i])
	if err != nil || pid <= 0 {
		return ptyServeInfo{}, false
	}
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	// ppid
	j := i
	for j < len(line) && line[j] >= '0' && line[j] <= '9' {
		j++
	}
	if j == i {
		return ptyServeInfo{}, false
	}
	ppid, err := strconv.Atoi(line[i:j])
	if err != nil {
		return ptyServeInfo{}, false
	}
	for j < len(line) && (line[j] == ' ' || line[j] == '\t') {
		j++
	}
	rest := strings.TrimSpace(line[j:])
	if rest == "" {
		return ptyServeInfo{PID: pid, PPID: ppid}, true
	}
	// Optional fields before command: etime, then state (single letter+flags like Ss, Z, R+).
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return ptyServeInfo{PID: pid, PPID: ppid}, true
	}
	etime := ""
	cmdStart := 0
	if looksLikeEtime(fields[0]) {
		etime = fields[0]
		cmdStart = 1
	}
	if cmdStart < len(fields) && looksLikePSState(fields[cmdStart]) {
		// Skip zombies entirely by returning empty command when state is Z.
		if strings.HasPrefix(fields[cmdStart], "Z") {
			return ptyServeInfo{PID: pid, PPID: ppid, Etime: etime, Command: "<defunct>"}, true
		}
		cmdStart++
	}
	cmdStr := strings.Join(fields[cmdStart:], " ")
	return ptyServeInfo{PID: pid, PPID: ppid, Etime: etime, Command: cmdStr}, true
}

func looksLikePSState(s string) bool {
	// ps state is short: R, S, D, Z, T, U, I plus optional flags (+, s, l, <, N, …)
	if s == "" || len(s) > 8 {
		return false
	}
	switch s[0] {
	case 'R', 'S', 'D', 'Z', 'T', 'U', 'I', 'W', 'K':
		// rest are flag chars
		for i := 1; i < len(s); i++ {
			c := s[i]
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '+' || c == '<' || c == 'N' {
				continue
			}
			return false
		}
		return true
	default:
		return false
	}
}

func looksLikeEtime(s string) bool {
	// Accept: 0:01, 01:23, 1-02:03:04, 123 (seconds-only rare)
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r >= '0' && r <= '9') || r == ':' || r == '-' {
			continue
		}
		return false
	}
	return true
}
