package procresolve

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	goproc "github.com/xhd2015/dot-pkgs/go-pkgs/proc"
)

// ListLiveProcs returns a best-effort process snapshot via shared go-pkgs/proc.
// On failure (ps missing or error), returns nil/empty.
func ListLiveProcs() []Proc {
	rows := goproc.List(goproc.Options{})
	if len(rows) == 0 {
		return nil
	}
	out := make([]Proc, len(rows))
	for i, p := range rows {
		out[i] = Proc{PID: p.PID, PPID: p.PPID, Cmd: p.Cmd}
	}
	return out
}

// LiveLsof returns open-file paths for pid via shared go-pkgs/proc OpenFiles.
// On failure returns nil (soft miss for that candidate).
func LiveLsof(pid int) []string {
	paths := goproc.OpenFiles(pid, goproc.Options{})
	if len(paths) == 0 {
		return nil
	}
	return paths
}

// LiveLsofMany returns open-file paths for many PIDs via one `lsof -p p1,p2,… -Fn`.
// Missing/empty PIDs are omitted. On total failure returns an empty map (soft).
func LiveLsofMany(pids []int) map[int][]string {
	out := map[int][]string{}
	if len(pids) == 0 {
		return out
	}
	seen := map[int]struct{}{}
	var uniq []int
	for _, pid := range pids {
		if pid <= 0 {
			continue
		}
		if _, ok := seen[pid]; ok {
			continue
		}
		seen[pid] = struct{}{}
		uniq = append(uniq, pid)
	}
	if len(uniq) == 0 {
		return out
	}
	if len(uniq) == 1 {
		paths := LiveLsof(uniq[0])
		if len(paths) > 0 {
			out[uniq[0]] = paths
		}
		return out
	}
	parts := make([]string, len(uniq))
	for i, pid := range uniq {
		parts[i] = strconv.Itoa(pid)
	}
	cmd := exec.Command("lsof", "-p", strings.Join(parts, ","), "-Fn")
	raw, err := cmd.Output()
	if err != nil && len(raw) == 0 {
		// Fall back to per-pid soft misses rather than failing the whole batch.
		for _, pid := range uniq {
			if paths := LiveLsof(pid); len(paths) > 0 {
				out[pid] = paths
			}
		}
		return out
	}
	return parseLsofFnByPID(raw)
}

// parseLsofFnByPID splits multi-process `lsof -Fn` output on `p<pid>` records.
func parseLsofFnByPID(raw []byte) map[int][]string {
	out := map[int][]string{}
	cur := -1
	seenPath := map[int]map[string]bool{}
	sc := bufio.NewScanner(bytes.NewReader(raw))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if len(line) < 2 {
			continue
		}
		switch line[0] {
		case 'p':
			pid, err := strconv.Atoi(line[1:])
			if err != nil {
				cur = -1
				continue
			}
			cur = pid
			if _, ok := out[cur]; !ok {
				out[cur] = nil
			}
			if _, ok := seenPath[cur]; !ok {
				seenPath[cur] = map[string]bool{}
			}
		case 'n':
			if cur <= 0 {
				continue
			}
			path := line[1:]
			if path == "" || path == "/" || !strings.HasPrefix(path, "/") {
				continue
			}
			if seenPath[cur][path] {
				continue
			}
			seenPath[cur][path] = true
			out[cur] = append(out[cur], path)
		}
	}
	return out
}

// FormatLiveError is a small helper for CLI messages (kept for symmetry).
func FormatLiveError(pid int, err error) string {
	return fmt.Sprintf("pid %d: %v", pid, err)
}
