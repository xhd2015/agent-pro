package ttywatch

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
	"time"
)

const (
	diagCmdTimeout   = 400 * time.Millisecond
	diagMaxAncestors = 16
	diagMaxChildren  = 12
)

// processInfo is a best-effort snapshot of a process for lock diagnostics.
type processInfo struct {
	PID     int
	PPID    int
	Command string
}

// formatRegistryLockBusyError builds a multi-line diagnostic error when the
// exclusive registry flock cannot be acquired within timeout.
//
// Includes summary, absolute lock path, flock holders (via lsof), process
// ancestry/children (via ps), and short remediation hints. Degrades gracefully
// when discovery tools fail. Kept cheap so total lock-timeout + diagnostics
// still fits short CLI/harness budgets (~3s).
func formatRegistryLockBusyError(lockPath string, timeout time.Duration) error {
	return fmt.Errorf("%s", formatRegistryLockBusyDiagnostics(lockPath, timeout))
}

func formatRegistryLockBusyDiagnostics(lockPath string, timeout time.Duration) string {
	absPath := lockPath
	if abs, err := filepath.Abs(lockPath); err == nil {
		absPath = abs
	}

	var b strings.Builder
	fmt.Fprintf(&b, "registry lock busy: timed out after %s waiting for exclusive flock\n", timeout)
	fmt.Fprintf(&b, "  lock:  %s\n", absPath)

	selfPID := os.Getpid()
	// One process-table snapshot for PPID/command + children (cheap under load).
	table := loadProcessTable()

	holders, listErr := listLockHolders(absPath, selfPID, table)
	if listErr != nil {
		fmt.Fprintf(&b, "\n  holders: could not list open-file holders (%v)\n", listErr)
	} else if len(holders) == 0 {
		fmt.Fprintf(&b, "\n  holders: none discovered (lsof reported no other openers)\n")
	} else {
		fmt.Fprintf(&b, "\n  holders (exclusive flock):\n")
		fmt.Fprintf(&b, "    %-7s %-7s %s\n", "PID", "PPID", "COMMAND")
		for _, h := range holders {
			fmt.Fprintf(&b, "    %-7d %-7d %s\n", h.PID, h.PPID, h.Command)
		}

		for _, h := range holders {
			fmt.Fprintf(&b, "\n  process tree (holder → root):\n")
			writeAncestorTree(&b, h, table)

			fmt.Fprintf(&b, "\n  children of holder (depth ≤2):\n")
			writeChildrenTree(&b, h, table)
		}
	}

	fmt.Fprintf(&b, "\n  what to do:\n")
	fmt.Fprintf(&b, "    - quit or stop the holding process tree, then retry\n")
	fmt.Fprintf(&b, "    - or: TTY_WATCH_HOME=$(mktemp -d) tty-watch run ...\n")
	fmt.Fprintf(&b, "    - (a concurrent tty-watch run is one possible holder)\n")

	// Trim trailing newline so CLI fmt.Fprintf(os.Stderr, "%s\n", err) adds exactly one.
	return strings.TrimRight(b.String(), "\n")
}

// listLockHolders returns processes (other than selfPID) that have lockPath open.
func listLockHolders(lockPath string, selfPID int, table map[int]processInfo) ([]processInfo, error) {
	pids, err := lsofPIDs(lockPath)
	if err != nil {
		return nil, err
	}
	seen := make(map[int]struct{}, len(pids))
	var out []processInfo
	for _, pid := range pids {
		if pid <= 0 || pid == selfPID {
			continue
		}
		if _, ok := seen[pid]; ok {
			continue
		}
		seen[pid] = struct{}{}
		info, ok := table[pid]
		if !ok {
			info = processInfo{PID: pid, Command: fmt.Sprintf("pid %d", pid)}
		}
		if info.Command == "" {
			info.Command = fmt.Sprintf("pid %d", pid)
		}
		out = append(out, info)
	}
	return out, nil
}

// lsofPIDs runs `lsof -F p -- <path>` and returns PIDs that have the path open.
func lsofPIDs(lockPath string) ([]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), diagCmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "lsof", "-F", "p", "--", lockPath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	// lsof exits 1 when no processes match; treat empty as no holders.
	if err != nil && stdout.Len() == 0 {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("lsof timed out")
		}
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("%s", msg)
	}
	var pids []int
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "p") {
			continue
		}
		pid, convErr := strconv.Atoi(line[1:])
		if convErr != nil || pid <= 0 {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

// loadProcessTable snapshots pid → {ppid, command} via a single `ps` invocation.
func loadProcessTable() map[int]processInfo {
	ctx, cancel := context.WithTimeout(context.Background(), diagCmdTimeout)
	defer cancel()
	// -axww: all processes, wide args (full command lines for marker matching).
	cmd := exec.CommandContext(ctx, "ps", "-axww", "-o", "pid=,ppid=,command=")
	out, err := cmd.Output()
	if err != nil {
		return map[int]processInfo{}
	}
	table := make(map[int]processInfo, 256)
	for _, line := range strings.Split(string(out), "\n") {
		info, ok := parsePSLine(line)
		if !ok {
			continue
		}
		table[info.PID] = info
	}
	return table
}

// parsePSLine parses a `ps -o pid=,ppid=,command=` line.
func parsePSLine(line string) (processInfo, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return processInfo{}, false
	}
	// pid
	i := 0
	for i < len(line) && line[i] >= '0' && line[i] <= '9' {
		i++
	}
	if i == 0 {
		return processInfo{}, false
	}
	pid, err := strconv.Atoi(line[:i])
	if err != nil || pid <= 0 {
		return processInfo{}, false
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
		return processInfo{}, false
	}
	ppid, err := strconv.Atoi(line[i:j])
	if err != nil {
		return processInfo{}, false
	}
	cmdStr := strings.TrimSpace(line[j:])
	return processInfo{PID: pid, PPID: ppid, Command: cmdStr}, true
}

func writeAncestorTree(b *strings.Builder, holder processInfo, table map[int]processInfo) {
	cmd := holder.Command
	if cmd == "" {
		cmd = fmt.Sprintf("pid %d", holder.PID)
	}
	fmt.Fprintf(b, "    %d  %s\n", holder.PID, cmd)
	seen := map[int]struct{}{holder.PID: {}}
	cur := holder.PPID
	indent := 1
	for cur > 0 && indent <= diagMaxAncestors {
		if _, ok := seen[cur]; ok {
			break
		}
		seen[cur] = struct{}{}
		info, ok := table[cur]
		cmd = ""
		nextPPID := 0
		if ok {
			cmd = info.Command
			nextPPID = info.PPID
		}
		if cmd == "" {
			cmd = fmt.Sprintf("pid %d", cur)
		}
		pad := strings.Repeat("  ", indent)
		fmt.Fprintf(b, "    %s└─ parent %d  %s\n", pad, cur, cmd)
		if cur == 1 {
			break
		}
		cur = nextPPID
		indent++
	}
}

func writeChildrenTree(b *strings.Builder, holder processInfo, table map[int]processInfo) {
	cmd := holder.Command
	if cmd == "" {
		cmd = fmt.Sprintf("pid %d", holder.PID)
	}
	fmt.Fprintf(b, "    %d  %s\n", holder.PID, cmd)

	children := childrenOf(holder.PID, table)
	if len(children) == 0 {
		fmt.Fprintf(b, "    (no children)\n")
		return
	}
	if len(children) > diagMaxChildren {
		children = children[:diagMaxChildren]
	}
	for _, child := range children {
		cCmd := child.Command
		if cCmd == "" {
			cCmd = fmt.Sprintf("pid %d", child.PID)
		}
		fmt.Fprintf(b, "      └─ %d  %s\n", child.PID, cCmd)
		grand := childrenOf(child.PID, table)
		if len(grand) > diagMaxChildren {
			grand = grand[:diagMaxChildren]
		}
		for _, g := range grand {
			gCmd := g.Command
			if gCmd == "" {
				gCmd = fmt.Sprintf("pid %d", g.PID)
			}
			fmt.Fprintf(b, "        └─ %d  %s\n", g.PID, gCmd)
		}
	}
}

func childrenOf(pid int, table map[int]processInfo) []processInfo {
	var kids []processInfo
	for _, info := range table {
		if info.PPID == pid {
			kids = append(kids, info)
		}
	}
	sort.Slice(kids, func(i, j int) bool { return kids[i].PID < kids[j].PID })
	return kids
}
