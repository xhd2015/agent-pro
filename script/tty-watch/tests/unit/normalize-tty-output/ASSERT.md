## Expected

- `yes\n[Terminal exited]\n` → `yes\r\n[Terminal exited]\r\n`
- `yes\r\n` unchanged
- `MARKER_A\rMARKER_B\n` → `MARKER_A\rMARKER_B\r\n`

```go
import (
	"bytes"
	"testing"

	"github.com/xhd2015/agent-pro/script/tty-watch/ttywatchtest"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "lf only adds cr for column reset",
			in:   "yes\n[Terminal exited]\n",
			want: "yes\r\n[Terminal exited]\r\n",
		},
		{
			name: "crlf unchanged",
			in:   "yes\r\n",
			want: "yes\r\n",
		},
		{
			name: "standalone cr preserved",
			in:   "MARKER_A\rMARKER_B\n",
			want: "MARKER_A\rMARKER_B\r\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ttywatchtest.NormalizeTTYOutput([]byte(tc.in))
			if !bytes.Equal(got, []byte(tc.want)) {
				t.Fatalf("NormalizeTTYOutput(%q) = %q, want %q", tc.in, got, tc.want)
			}
			if ttywatchtest.HasBareLFNewlines(string(got)) {
				t.Fatalf("normalized output has bare LF newlines: %q", got)
			}
		})
	}
}
```