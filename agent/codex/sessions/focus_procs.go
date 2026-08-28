package sessions

import (
	"bufio"
	"bytes"
	"os/exec"
	"strconv"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// FocusProc is one process-table row used to walk from a live codex PID to a TTY.
type FocusProc struct {
	PID  int
	PPID int
	TTY  string // e.g. /dev/ttys148, ttys148, "??", or ""
	Cmd  string
}

func listLiveFocusProcs() []FocusProc {
	cmd := exec.Command("ps", "-ax", "-o", "pid=,ppid=,tty=,command=")
	out, err := cmd.Output()
	if err != nil {
		cmd = exec.Command("ps", "-x", "-o", "pid=,ppid=,tty=,command=")
		out, err = cmd.Output()
		if err != nil {
			return nil
		}
	}
	return parseFocusProcRows(out)
}

func parseFocusProcRows(out []byte) []FocusProc {
	var rows []FocusProc
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		pid, err1 := strconv.Atoi(fields[0])
		ppid, err2 := strconv.Atoi(fields[1])
		if err1 != nil || err2 != nil {
			continue
		}
		cmdStr := ""
		if len(fields) > 3 {
			cmdStr = strings.Join(fields[3:], " ")
		}
		rows = append(rows, FocusProc{PID: pid, PPID: ppid, TTY: fields[2], Cmd: cmdStr})
	}
	return rows
}

// collectTTYsFromTree returns real TTYs from the root PID plus its ancestors
// and descendants. Skips "??", blank, and whitespace-only TTY fields.
func collectTTYsFromTree(procs []FocusProc, rootPID int) []string {
	if rootPID <= 0 || len(procs) == 0 {
		return nil
	}
	byPID := make(map[int]FocusProc, len(procs))
	children := make(map[int][]int, len(procs))
	for _, p := range procs {
		byPID[p.PID] = p
		if p.PPID == 0 && p.PID == p.PPID {
			continue
		}
		children[p.PPID] = append(children[p.PPID], p.PID)
	}
	if _, ok := byPID[rootPID]; !ok {
		return nil
	}

	seenPID := map[int]bool{rootPID: true}
	var ordered []FocusProc

	cur := rootPID
	var ancestors []FocusProc
	for {
		row, ok := byPID[cur]
		if !ok {
			break
		}
		if cur != rootPID {
			ancestors = append(ancestors, row)
			seenPID[cur] = true
		}
		if row.PPID == 0 || row.PPID == cur {
			break
		}
		if seenPID[row.PPID] {
			break
		}
		cur = row.PPID
	}

	ordered = append(ordered, byPID[rootPID])
	ordered = append(ordered, ancestors...)

	queue := []int{rootPID}
	for len(queue) > 0 {
		pid := queue[0]
		queue = queue[1:]
		for _, child := range children[pid] {
			if seenPID[child] {
				continue
			}
			seenPID[child] = true
			if row, ok := byPID[child]; ok {
				ordered = append(ordered, row)
			}
			queue = append(queue, child)
		}
	}

	seenTTY := map[string]bool{}
	var out []string
	for _, row := range ordered {
		tty := strings.TrimSpace(row.TTY)
		if tty == "" || tty == "??" {
			continue
		}
		norm := iterm2.NormalizeTTY(tty)
		if norm == "" || seenTTY[norm] {
			continue
		}
		seenTTY[norm] = true
		out = append(out, norm)
	}
	return out
}
