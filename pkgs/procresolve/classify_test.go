package procresolve

import "testing"

func TestIsCodexRunner(t *testing.T) {
	t.Parallel()
	cases := []struct {
		cmd  string
		want bool
	}{
		{"/usr/local/bin/codex", true},
		{"codex resume abc", true},
		{"/usr/local/bin/grok", false},
		{"/bin/bash", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := IsCodexRunner(tc.cmd); got != tc.want {
			t.Fatalf("IsCodexRunner(%q)=%v want %v", tc.cmd, got, tc.want)
		}
	}
}

func TestFindAncestorCodex(t *testing.T) {
	t.Parallel()
	procs := []Proc{
		{PID: 1, PPID: 0, Cmd: "/sbin/launchd"},
		{PID: 100, PPID: 1, Cmd: "/usr/local/bin/codex"},
		{PID: 200, PPID: 100, Cmd: "/bin/bash"},
		{PID: 300, PPID: 200, Cmd: "/usr/local/bin/agent-pro"},
		{PID: 400, PPID: 1, Cmd: "/usr/local/bin/grok"},
		{PID: 500, PPID: 400, Cmd: "/usr/local/bin/agent-pro"},
	}
	opts := Options{
		ListProcs: func() []Proc { return procs },
	}
	anc, ok := FindAncestorCodex(300, opts)
	if !ok || anc.PID != 100 {
		t.Fatalf("FindAncestorCodex(300)=(%+v,%v) want pid 100", anc, ok)
	}
	if _, ok := FindAncestorCodex(500, opts); ok {
		t.Fatalf("FindAncestorCodex on grok chain must miss")
	}
}
