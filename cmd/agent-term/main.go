package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
	ptyclient "github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap/client"
	"github.com/xhd2015/less-gen/flags"
	"golang.org/x/term"
)

const defaultListen = "127.0.0.1:7681"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "\n%s\n", err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: agent-term <serve|list|run|attach|rename|web>")
	}
	switch args[0] {
	case "serve":
		return runServe(args[1:])
	case "list":
		return runList(args[1:])
	case "run":
		return runRun(args[1:])
	case "attach":
		return runAttach(args[1:])
	case "rename":
		return runRename(args[1:])
	case "web":
		return runWeb(args[1:])
	case "-h", "--help", "help":
		printHelp()
		return nil
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printHelp() {
	fmt.Print(`agent-term — local terminal session daemon

Commands:
  serve [--listen ADDR]   Start TCP daemon (default 127.0.0.1:7681)
  list                    List sessions
  run <cmd> [args...]     Create session, attach, print id on exit
  attach <id-or-name>     Attach to existing session
  rename <id-or-name> <name>
  web <id-or-name> [--port PORT]
`)
}

func runServe(args []string) error {
	listen := defaultListen
	args, err := flags.String("--listen", &listen).Parse(args)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		return fmt.Errorf("serve: unexpected arguments %v", args)
	}
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "\nlistening on %s\n%s\n", listen, listen)
	mux := http.NewServeMux()
	ptywrap.RegisterAPI(mux)
	srv := &http.Server{Handler: loggingMiddleware(mux)}
	return srv.Serve(ln)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if isWebSocketUpgrade(r) && path == "/api/terminal" {
			sessionID := r.URL.Query().Get("session_id")
			if sessionID != "" {
				fmt.Fprintf(os.Stderr, "agent-term: WS /api/terminal session_id=%s\n", sessionID)
			} else {
				fmt.Fprintf(os.Stderr, "agent-term: WS /api/terminal\n")
			}
		} else {
			fmt.Fprintf(os.Stderr, "%s %s\n%s\n", r.Method, path, path)
		}
		next.ServeHTTP(w, r)
	})
}

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

func runList(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("list: unexpected arguments %v", args)
	}
	c := ptyclient.NewDefaultClient()
	sessions, err := c.List()
	if err != nil {
		return err
	}
	if len(sessions) == 0 {
		fmt.Println("No terminal sessions.")
		return nil
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(sessions)
}

func runRun(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("run: command required")
	}
	c := ptyclient.NewDefaultClient()
	info, err := c.Create(args, "", "")
	if err != nil {
		return err
	}
	if isInteractiveTerminal(os.Stdin, os.Stdout) {
		// Fast commands (true, echo, …) often exit before WS attach connects.
		// Brief poll avoids hanging in attach on a session that already exited
		// (stdin forward would block forever without a Terminal-exited frame).
		for i := 0; i < 15; i++ {
			if sessionExited(c, info.ID) {
				fmt.Printf("\n%s\n", info.ID)
				return nil
			}
			time.Sleep(20 * time.Millisecond)
		}
		attachErr := attachSession(c, info.ID)
		if errors.Is(attachErr, errDetached) {
			return attachErr
		}
		if attachErr != nil && !sessionExited(c, info.ID) {
			return attachErr
		}
		fmt.Printf("\n%s\n", info.ID)
		return nil
	}
	if commandNeedsInteractiveTTY(args) {
		return fmt.Errorf("interactive terminal required on stdin/stdout\nterminal attach requires a TTY (same as attach)")
	}
	if err := ptyclient.WaitSession(c, info.ID); err != nil {
		return err
	}
	fmt.Println(info.ID)
	return nil
}

func isInteractiveTerminal(stdin io.Reader, stdout io.Writer) bool {
	in, okIn := stdin.(*os.File)
	out, okOut := stdout.(*os.File)
	if !okIn || !okOut {
		return false
	}
	if !term.IsTerminal(int(in.Fd())) || !term.IsTerminal(int(out.Fd())) {
		return false
	}
	if st, err := in.Stat(); err == nil && st.Mode()&os.ModeNamedPipe != 0 {
		return false
	}
	if st, err := out.Stat(); err == nil && st.Mode()&os.ModeNamedPipe != 0 {
		return false
	}
	return true
}

func commandNeedsInteractiveTTY(args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch filepath.Base(args[0]) {
	case "bash", "zsh", "fish", "grok":
		return !hasShellScriptArg(args[1:])
	case "sh":
		return len(args) == 1 || !hasShellScriptArg(args[1:])
	default:
		return false
	}
}

func sessionExited(c *ptyclient.Client, sessionID string) bool {
	sessions, err := c.List()
	if err != nil {
		return false
	}
	for _, s := range sessions {
		if s.ID == sessionID && s.Status == "exited" {
			return true
		}
	}
	return false
}

func hasShellScriptArg(args []string) bool {
	for i, arg := range args {
		if arg == "-c" && i+1 < len(args) {
			return true
		}
	}
	return false
}

func runAttach(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("attach: requires <id-or-name>")
	}
	c := ptyclient.NewDefaultClient()
	session, err := ptyclient.ResolveTarget(c, args[0])
	if err != nil {
		return err
	}
	return attachSession(c, session.ID)
}

func runRename(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("rename: requires <id-or-name> <name>")
	}
	c := ptyclient.NewDefaultClient()
	session, err := ptyclient.ResolveTarget(c, args[0])
	if err != nil {
		return err
	}
	return c.Rename(session.ID, args[1])
}

func runWeb(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("web: requires <id-or-name>")
	}
	target := args[0]
	port := 0
	rest, err := flags.Int("--port", &port).Parse(args[1:])
	if err != nil {
		return err
	}
	if len(rest) > 0 {
		return fmt.Errorf("web: unexpected arguments %v", rest)
	}

	c := ptyclient.NewDefaultClient()
	session, err := ptyclient.ResolveTarget(c, target)
	if err != nil {
		return err
	}

	daemonWS := websocketURL(c.BaseURL, session.ID)
	page := webPageHTML(daemonWS, session.ID)

	var ln net.Listener
	if port > 0 {
		ln, err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	} else {
		ln, err = net.Listen("tcp", "127.0.0.1:0")
	}
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, page)
	})
	fmt.Fprintf(os.Stderr, "agent-term web listening on http://%s/\n", ln.Addr().String())
	return http.Serve(ln, mux)
}

func websocketURL(base, sessionID string) string {
	base = strings.TrimRight(base, "/")
	switch {
	case strings.HasPrefix(base, "https://"):
		base = "wss://" + strings.TrimPrefix(base, "https://")
	default:
		base = strings.TrimPrefix(base, "http://")
		base = "ws://" + base
	}
	return base + "/api/terminal?session_id=" + url.QueryEscape(sessionID)
}

func webPageHTML(wsURL, sessionID string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>agent-term %s</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.css" />
  <style>html,body{height:100%%;margin:0;background:#1e1e1e}#terminal{height:100%%;width:100%%}</style>
</head>
<body>
  <div id="terminal"></div>
  <script src="https://cdn.jsdelivr.net/npm/xterm@5.3.0/lib/xterm.js"></script>
  <script>
    const term = new Terminal({cursorBlink:true, fontSize:14, theme:{background:'#1e1e1e'}});
    term.open(document.getElementById('terminal'));
    term.writeln('Connecting to %s ...');
    const ws = new WebSocket(%q);
    ws.binaryType = 'arraybuffer';
    ws.onmessage = (ev) => {
      if (typeof ev.data === 'string') {
        try {
          const msg = JSON.parse(ev.data);
          if (msg.type === 'session_id') return;
        } catch (_) {}
        term.write(ev.data);
      } else {
        term.write(new Uint8Array(ev.data));
      }
    };
    term.onData((data) => ws.send(new TextEncoder().encode(data)));
    window.addEventListener('resize', () => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({type:'resize', cols:term.cols, rows:term.rows}));
      }
    });
  </script>
</body>
</html>`, sessionID, sessionID, wsURL)
}