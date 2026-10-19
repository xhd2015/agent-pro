package sessions

import (
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
)

func TestListLiveOpenLockOnly(t *testing.T) {
	const (
		liveA = "session-4a7802d3-18e0-4fd3-a7ef-a45447a78b81"
		liveB = "session-c7e136af-8667-4b6a-bf68-4b9d17cca85b"
	)
	lock := func(id string) string {
		return "/Users/x/.dsh/sessions/--proj--/" + id + "/session.lock"
	}
	log := "/Users/x/.dsh/sessions/--proj--/" + liveA + "/session.v3.jsonl.zstd"

	rows, err := ListLive(&ListLiveOpts{
		ListProcs: func() []procresolve.Proc {
			return []procresolve.Proc{
				{PID: 10, Cmd: "node --import tsx/esm apps/cli/src/bin.ts web --no-open"},
				{PID: 11, Cmd: "node server.js"},
				{PID: 12, Cmd: "/usr/local/bin/dsh tui"},
			}
		},
		Lsof: func(pid int) []string {
			switch pid {
			case 10:
				return []string{lock(liveA), lock(liveB), log}
			case 12:
				return []string{lock(liveA)}
			default:
				return []string{lock(liveB)}
			}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows=%d (%+v), want 2", len(rows), rows)
	}
	if rows[0].ID != liveA || rows[1].ID != liveB {
		t.Fatalf("ids = %s %s", rows[0].ID, rows[1].ID)
	}
	if rows[0].PID != 10 || rows[1].PID != 10 {
		t.Fatalf("pids = %d %d, want the web server for both", rows[0].PID, rows[1].PID)
	}
}

func TestListLiveIgnoresALockFileWithNoOpenFD(t *testing.T) {
	rows, err := ListLive(&ListLiveOpts{
		ListProcs: func() []procresolve.Proc {
			return []procresolve.Proc{{PID: 10, Cmd: "dsh web"}}
		},
		Lsof: func(int) []string { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("rows=%+v, want none", rows)
	}
}
