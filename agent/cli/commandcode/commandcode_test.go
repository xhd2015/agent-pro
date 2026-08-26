package commandcode

import "testing"

func TestParseCommandcodeJSONLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		line      string
		wantText  string
		wantSess  string
	}{
		{
			name:     "run_start event captures session",
			line:     `{"type":"event","event":{"type":"run_start","sessionId":"sess-1"}}`,
			wantText: "",
			wantSess: "sess-1",
		},
		{
			name:     "result carries finalText and session",
			line:     `{"type":"result","subtype":"success","sessionId":"sess-2","finalText":"hello world"}`,
			wantText: "hello world",
			wantSess: "sess-2",
		},
		{
			name:     "non-json ignored",
			line:     `not-json`,
			wantText: "",
			wantSess: "",
		},
		{
			name:     "empty finalText still yields session",
			line:     `{"type":"result","sessionId":"sess-3","finalText":""}`,
			wantText: "",
			wantSess: "sess-3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			text, sess := parseCommandcodeJSONLine(tt.line)
			if text != tt.wantText {
				t.Fatalf("text = %q, want %q", text, tt.wantText)
			}
			if sess != tt.wantSess {
				t.Fatalf("session = %q, want %q", sess, tt.wantSess)
			}
		})
	}
}
