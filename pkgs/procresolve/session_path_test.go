package procresolve

import "testing"

func TestGrokSessionDirFromPath(t *testing.T) {
	path := "/Users/x/.grok/sessions/%2Ftmp%2Fproj/019f283a-aaaa-7aaa-aaaa-aaaaaaaaaa01/events.jsonl"
	dir, sid, ok := GrokSessionDirFromPath(path)
	if !ok {
		t.Fatal("expected ok")
	}
	if sid != "019f283a-aaaa-7aaa-aaaa-aaaaaaaaaa01" {
		t.Fatalf("sid=%q", sid)
	}
	wantDir := "/Users/x/.grok/sessions/%2Ftmp%2Fproj/019f283a-aaaa-7aaa-aaaa-aaaaaaaaaa01"
	if dir != wantDir {
		t.Fatalf("dir=%q want %q", dir, wantDir)
	}
}

func TestParseLsofFnByPID(t *testing.T) {
	raw := []byte("p100\nf1\nn/tmp/a\np200\nn/tmp/b\nn/tmp/a\n")
	got := parseLsofFnByPID(raw)
	if len(got[100]) != 1 || got[100][0] != "/tmp/a" {
		t.Fatalf("pid 100: %#v", got[100])
	}
	if len(got[200]) != 2 {
		t.Fatalf("pid 200: %#v", got[200])
	}
}
