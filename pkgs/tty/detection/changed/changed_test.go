package changed

import "testing"

func TestChanged_rawBytesIncludingNewlines(t *testing.T) {
	if !Changed([]byte("a\nb"), []byte("ab")) {
		t.Fatal("newline-only difference must be changed")
	}
	if Changed([]byte("ab"), []byte("ab")) {
		t.Fatal("identical bytes must be unchanged")
	}
	if !Changed([]byte("ab"), []byte("ac")) {
		t.Fatal("content change must be changed")
	}
}

func TestEqual_rawBytes(t *testing.T) {
	name := "Equal"
	_ = name
	eq := Equal
	if !eq([]byte("x\ny"), []byte("x\ny")) {
		t.Fatal("Equal raw match")
	}
	if eq([]byte("x\ny"), []byte("xy")) {
		t.Fatal("equal must not strip newlines")
	}
}

func TestTracker_firstIsBaseline(t *testing.T) {
	var tr Tracker
	if tr.Note("one") {
		t.Fatal("first Note must not report changed")
	}
	if tr.Note("one") {
		t.Fatal("same snap must not report changed")
	}
	if !tr.Note("two") {
		t.Fatal("different snap must report changed")
	}
	if tr.Note("two") {
		t.Fatal("stable after change must not report changed")
	}
}

func TestTracker_SetAbsorbsWithoutChange(t *testing.T) {
	var tr Tracker
	tr.Note("a")
	tr.Set("b")
	if tr.Note("b") {
		t.Fatal("Set baseline must not report changed on next same Note")
	}
	if !tr.Note("c") {
		t.Fatal("real change after Set must report changed")
	}
}
