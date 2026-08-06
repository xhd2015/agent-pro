package procresolve

import (
	"fmt"

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

// FormatLiveError is a small helper for CLI messages (kept for symmetry).
func FormatLiveError(pid int, err error) string {
	return fmt.Sprintf("pid %d: %v", pid, err)
}
