package sessions

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
)

func pidAlive(pid int) bool {
	if pid <= 1 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}

func allPIDsDead(pids []int) bool {
	for _, pid := range pids {
		if pidAlive(pid) {
			return false
		}
	}
	return true
}

func uniquePositivePIDs(pids []int) []int {
	seen := make(map[int]bool, len(pids))
	var out []int
	for _, pid := range pids {
		if pid <= 1 || seen[pid] {
			continue
		}
		seen[pid] = true
		out = append(out, pid)
	}
	return out
}

// readinessWatchPIDs returns lock-holder PIDs plus their parent chain so
// closing the iTerm window (which kills agent-run run while __serve may
// briefly linger) still wakes wait via NOTE_EXIT.
//
// Parent walk uses one-shot `ps -o ppid=` per pid (not ListLiveProcs), so we
// still see agent-run run / login even if a filtered proc list omits them.
func readinessWatchPIDs(lockPath string, listProcs func() []procresolve.Proc) []int {
	_ = listProcs // kept for WaitOpts.Live injection symmetry; unused
	holders := pidsHoldingPath(lockPath)
	if len(holders) == 0 {
		return nil
	}

	var out []int
	seen := map[int]bool{}
	add := func(pid int) {
		if pid <= 1 || seen[pid] {
			return
		}
		seen[pid] = true
		out = append(out, pid)
	}
	for _, h := range holders {
		add(h)
		pid := h
		for i := 0; i < 10; i++ {
			ppid := parentPID(pid)
			if ppid <= 1 {
				break
			}
			add(ppid)
			pid = ppid
		}
	}
	return out
}

func parentPID(pid int) int {
	out, err := exec.Command("ps", "-o", "ppid=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return 0
	}
	s := strings.TrimSpace(string(out))
	if s == "" {
		return 0
	}
	ppid, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return ppid
}

func pidsHoldingPath(path string) []int {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	// One-shot lsof; not a poll loop.
	out, err := exec.Command("lsof", "-t", path).Output()
	if err != nil {
		return nil
	}
	var pids []int
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err != nil || pid <= 1 {
			continue
		}
		pids = append(pids, pid)
	}
	return pids
}
