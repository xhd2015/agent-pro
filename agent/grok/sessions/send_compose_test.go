package sessions

import (
	"testing"

	lessflags "github.com/xhd2015/less-flags"
)

func TestComposeSendPayload_Order(t *testing.T) {
	seq := lessflags.Flags{
		{Flag: "--up"},
		{Flag: "--text", Value: "pick"},
		{Flag: "--enter"},
	}
	got, hasText, hasEnter, err := composeSendPayload(seq, "tail")
	if err != nil {
		t.Fatal(err)
	}
	want := "\x1b[Apick\ntail"
	if got != want {
		t.Fatalf("payload=%q want %q", got, want)
	}
	if !hasText || !hasEnter {
		t.Fatalf("hasText=%v hasEnter=%v", hasText, hasEnter)
	}
}

func TestComposeSendPayload_KeyOnly(t *testing.T) {
	seq := lessflags.Flags{{Flag: "--ctrl-c"}, {Flag: "--ctrl-c"}}
	got, hasText, hasEnter, err := composeSendPayload(seq, "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "\x03\x03" || hasText || hasEnter {
		t.Fatalf("got %q hasText=%v hasEnter=%v", got, hasText, hasEnter)
	}
}

func TestComposeSendPayload_PositionalOnly(t *testing.T) {
	got, hasText, hasEnter, err := composeSendPayload(nil, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" || !hasText || hasEnter {
		t.Fatalf("got %q hasText=%v hasEnter=%v", got, hasText, hasEnter)
	}
}

func TestParseSendArgs_Sequence(t *testing.T) {
	out, err := parseSendArgs([]string{
		"--up", "--text", "a", "--enter", "b",
		"--session-id", "s1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Text != "b" {
		t.Fatalf("Text=%q", out.Text)
	}
	if out.SessionID == nil || *out.SessionID != "s1" {
		t.Fatalf("SessionID=%v", out.SessionID)
	}
	want := lessflags.Flags{
		{Flag: "--up", Value: ""},
		{Flag: "--text", Value: "a"},
		{Flag: "--enter", Value: ""},
	}
	if len(out.Seq) != len(want) {
		t.Fatalf("Seq=%#v", out.Seq)
	}
	for i := range want {
		if out.Seq[i] != want[i] {
			t.Fatalf("Seq[%d]=%#v want %#v", i, out.Seq[i], want[i])
		}
	}
}
