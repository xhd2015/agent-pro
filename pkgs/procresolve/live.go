package procresolve

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// ListLiveProcs returns a best-effort process snapshot via `ps`.
// On failure (ps missing or error), returns nil.
func ListLiveProcs() []Proc {
	// macOS/BSD and Linux: pid, ppid, full command line.
	cmd := exec.Command("ps", "-ax", "-o", "pid=,ppid=,command=")
	out, err := cmd.Output()
	if err != nil {
		// Try without -a (some environments).
		cmd = exec.Command("ps", "-x", "-o", "pid=,ppid=,command=")
		out, err = cmd.Output()
		if err != nil {
			return nil
		}
	}
	return parsePSOutput(out)
}

func parsePSOutput(out []byte) []Proc {
	var procs []Proc
	sc := bufio.NewScanner(bytes.NewReader(out))
	// Long command lines.
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		// pid and ppid are leading integers; remainder is command.
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pid, err1 := strconv.Atoi(fields[0])
		ppid, err2 := strconv.Atoi(fields[1])
		if err1 != nil || err2 != nil {
			continue
		}
		cmd := ""
		if len(fields) > 2 {
			// Rejoin from original after the two numeric fields.
			// fields-based join loses multiple spaces; acceptable for classify.
			cmd = strings.Join(fields[2:], " ")
		}
		procs = append(procs, Proc{PID: pid, PPID: ppid, Cmd: cmd})
	}
	return procs
}

// LiveLsof returns open-file paths for pid via `lsof -p <pid> -Fn`.
// On failure returns nil (soft miss for that candidate).
func LiveLsof(pid int) []string {
	if pid <= 0 {
		return nil
	}
	cmd := exec.Command("lsof", "-p", strconv.Itoa(pid), "-Fn")
	out, err := cmd.Output()
	if err != nil {
		// lsof exits non-zero when no files / no process; still may have stdout.
		if len(out) == 0 {
			return nil
		}
	}
	return parseLsofFn(out)
}

func parseLsofFn(out []byte) []string {
	var paths []string
	seen := map[string]bool{}
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if len(line) < 2 || line[0] != 'n' {
			continue
		}
		path := line[1:]
		if path == "" || path == "/" {
			continue
		}
		// Skip pseudo paths like "txt" descriptors already filtered by -Fn name field.
		if strings.HasPrefix(path, "/") && !seen[path] {
			seen[path] = true
			paths = append(paths, path)
		}
	}
	return paths
}

// FormatLiveError is a small helper for CLI messages (kept for symmetry).
func FormatLiveError(pid int, err error) string {
	return fmt.Sprintf("pid %d: %v", pid, err)
}
