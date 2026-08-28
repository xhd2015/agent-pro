package sessions

import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/procresolve"
)

func TestStatus_Running(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff61"
	writeCodexRollout(t, home, sid)

	live := &LiveOptions{
		ListProcs: func() []procresolve.Proc {
			return []procresolve.Proc{{PID: 4242, PPID: 1, Cmd: "/usr/local/bin/codex"}}
		},
		Lsof: func(pid int) []string {
			if pid != 4242 {
				return nil
			}
			// Hard-hit path must contain /.codex/sessions/…/rollout-*-<uuid>
			return []string{codexRolloutPath(sid)}
		},
	}
	st, err := Status(home, sid, true, live)
	if err != nil {
		t.Fatal(err)
	}
	if st.State != "running" {
		t.Fatalf("state=%q", st.State)
	}
	if st.FileActive {
		t.Fatal("FileActive must be false for Codex")
	}
	if len(st.PIDs) != 1 || st.PIDs[0].PID != 4242 {
		t.Fatalf("pids=%v", st.PIDs)
	}
	text := FormatStatusText(st)
	if !strings.Contains(text, "State: running") || !strings.Contains(text, "File: no") {
		t.Fatalf("text:\n%s", text)
	}
	if !strings.Contains(text, "4242") {
		t.Fatalf("want pid in text:\n%s", text)
	}
}

func TestStatus_Inactive(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff62"
	writeCodexRollout(t, home, sid)
	st, err := Status(home, sid, true, &LiveOptions{
		ListProcs: func() []procresolve.Proc { return nil },
		Lsof:      func(int) []string { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if st.State != "inactive" {
		t.Fatalf("state=%q", st.State)
	}
	if st.FileActive {
		t.Fatal("FileActive must be false")
	}
	text := FormatStatusText(st)
	if !strings.Contains(text, "PIDs: none") {
		t.Fatalf("text:\n%s", text)
	}
}

func TestStatus_NoPID(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff63"
	writeCodexRollout(t, home, sid)
	st, err := Status(home, sid, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if st.PIDChecked {
		t.Fatal("PIDChecked")
	}
	if st.State != "inactive" {
		t.Fatalf("state=%q want inactive without file signal", st.State)
	}
	text := FormatStatusText(st)
	if !strings.Contains(text, "skipped") {
		t.Fatalf("text:\n%s", text)
	}
}

func TestStatus_Unknown(t *testing.T) {
	t.Parallel()
	_, err := Status(t.TempDir(), "019f283a-ffff-7fff-ffff-ffffffffff99", true, nil)
	if err == nil || !strings.Contains(err.Error(), "codex session not found") {
		t.Fatalf("err=%v", err)
	}
}

func TestFormatStatusJSON_AbsolutePath(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	sid := "019f283a-ffff-7fff-ffff-ffffffffff64"
	writeCodexRollout(t, home, sid)
	st, err := Status(home, sid, true, &LiveOptions{
		ListProcs: func() []procresolve.Proc { return nil },
		Lsof:      func(int) []string { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := FormatStatusJSON(st)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatal("ANSI in JSON")
	}
	if !strings.Contains(out, `"file_active": false`) {
		t.Fatalf("want file_active false:\n%s", out)
	}
	if !strings.Contains(out, "rollout-") || strings.Contains(out, `"path": "~`) {
		t.Fatalf("want absolute rollout path:\n%s", out)
	}
}

func TestFormatActiveBlock(t *testing.T) {
	t.Parallel()
	block := FormatActiveBlock(&SessionStatus{
		FileActive: false,
		PIDChecked: true,
		PIDs:       []LivePID{{PID: 7, Name: "codex"}},
	})
	if !strings.Contains(block, "Active:") || !strings.Contains(block, "File: no") {
		t.Fatalf("%s", block)
	}
	if !strings.Contains(block, "7 codex") {
		t.Fatalf("%s", block)
	}
}
