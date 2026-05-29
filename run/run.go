package run

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xhd2015/agent-pro/server"
	"github.com/xhd2015/agent-pro/trace"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: agent-traces [OPTIONS]

Start the standalone agent trace viewer for agent runs stored under
~/.knowledge-hub/agent-traces.

Options:
  --dev                    run in dev mode (proxies to the vite dev server)
  --port PORT              listen on PORT (default: 9898)
  --workspace DIR          workspace directory (default: cwd)
  --data-dir DIR           portal data dir (default: ~/.knowledge-hub)
  --route-prefix PREFIX    mount the trace viewer under PREFIX, e.g. agent-traces
  --component NAME         render a single named component (default: full app)
  --open                   open the trace viewer in a browser after startup
  --no-open                do not open a browser
  -h, --help               show this help
`

func Run(args []string) error {
	var devFlag bool
	var component string
	var port int
	var workspace string
	var dataDir string
	var routePrefix string
	var openBrowser bool
	var noOpen bool
	args, err := flags.
		Bool("--dev", &devFlag).
		Int("--port", &port).
		String("--workspace", &workspace).
		String("--data-dir", &dataDir).
		String("--route-prefix", &routePrefix).
		String("--component", &component).
		Bool("--open", &openBrowser).
		Bool("--no-open", &noOpen).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra args: %s", strings.Join(args, " "))
	}

	if workspace == "" {
		workspace, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	_ = workspace

	if dataDir == "" {
		dataDir, err = defaultKnowledgeHubDir()
		if err != nil {
			return err
		}
	}
	if noOpen {
		openBrowser = false
	}
	server.SetDataDir(dataDir)

	if component == "list" {
		fmt.Println("Available components: App")
		return nil
	}

	if port == 0 {
		port = 9898
		if !server.CheckPortAvailable(port) {
			port, err = server.FindAvailablePort(9899, 100)
			if err != nil {
				return err
			}
		}
	}

	traceRoot, err := trace.AgentTraceRootForDataDir(dataDir)
	if err != nil {
		return err
	}
	fmt.Printf("Agent traces root: %s\n", traceRoot)

	if openBrowser {
		go openLocalURL(port, routePrefix)
	}

	if component != "" {
		var html string
		if !devFlag {
			html, err = server.FormatTemplateHtml(server.FormatOptions{
				Component: component,
			})
			if err != nil {
				return err
			}
		}
		return server.ServeComponent(port, server.ServeOptions{
			Dev:         devFlag,
			RoutePrefix: routePrefix,
			Static: server.StaticOptions{
				IndexHtml: html,
			},
		})
	}

	return server.Serve(port, devFlag, routePrefix)
}

func defaultKnowledgeHubDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".knowledge-hub"), nil
}

func openLocalURL(port int, routePrefix string) {
	url := fmt.Sprintf("http://localhost:%d%s", port, server.JoinRoutePrefix(routePrefix, "/"))
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
