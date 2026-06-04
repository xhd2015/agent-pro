package run

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/xhd2015/agent-pro/server"
	"github.com/xhd2015/agent-pro/trace"

	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: agent-traces [OPTIONS] [SOURCE...]

Start the standalone agent trace viewer. With no SOURCE, agent-traces
discovers trace roots from ~/.agent-traces, ~/.*/agent-traces,
./.agent-traces, and ./.*/agent-traces.

Options:
  --dev                    run in dev mode (proxies to the vite dev server)
  --port PORT              listen on PORT (default: 9898)
  --workspace DIR          workspace directory (default: cwd)
  --data-dir DIR           portal data dir; reads DIR/agent-traces
  --static-dir DIR         frontend dist directory to serve instead of embedded dist
  --focus-command COMMAND  foreground traces for a command, e.g. murphy or codenn
  --include-linked         include linked parent/child traces with focused traces
  --print                  print trace summary to terminal and exit
  --print-messages N       number of recent normalized messages to print (default: 3)
  --route-prefix PREFIX    mount the trace viewer under PREFIX, e.g. agent-traces
  --component NAME         render a single named component (default: full app)
  --open                   open the trace viewer in a browser after startup
  --no-open                do not open a browser
  -h, --help               show this help

SOURCE:
  file                      read a single JSONL agent-events file
  trace session directory   read one trace directory
  trace root directory      read multiple trace session directories
  config directory          recursively discover nested agent-traces roots
`

func Run(args []string) error {
	var devFlag bool
	var component string
	var port int
	var workspace string
	var dataDir string
	var staticDir string
	var focusCommand string
	var includeLinked bool
	var noPrintSources bool
	var printMode bool
	var printMessages int
	var routePrefix string
	var openBrowser bool
	var noOpen bool
	args, err := flags.
		Bool("--dev", &devFlag).
		Int("--port", &port).
		String("--workspace", &workspace).
		String("--data-dir", &dataDir).
		String("--static-dir", &staticDir).
		String("--focus-command", &focusCommand).
		Bool("--include-linked", &includeLinked).
		Bool("--no-print-sources", &noPrintSources).
		Bool("--print", &printMode).
		Int("--print-messages", &printMessages).
		String("--route-prefix", &routePrefix).
		String("--component", &component).
		Bool("--open", &openBrowser).
		Bool("--no-open", &noOpen).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if workspace == "" {
		workspace, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	_ = workspace

	if noOpen {
		openBrowser = false
	}

	if component == "list" {
		fmt.Println("Available components: App")
		return nil
	}

	source, sourceDescriptions, err := resolveTraceSource(dataDir, workspace, args)
	if err != nil {
		return err
	}
	source = trace.NewFocusSource(source, focusCommand, includeLinked)
	server.SetTraceSource(source)
	if staticDir != "" {
		server.SetFrontendDistDir(staticDir)
	}

	if printMode {
		if printMessages == 0 {
			printMessages = 3
		}
		return printTraceReport(source, sourceDescriptions, printMessages)
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

	if !noPrintSources {
		printTraceSources(sourceDescriptions)
	}

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

func resolveTraceSource(dataDir, workspace string, args []string) (trace.Source, []string, error) {
	if dataDir != "" {
		if len(args) > 0 {
			return nil, nil, fmt.Errorf("--data-dir cannot be combined with source arguments: %s", strings.Join(args, " "))
		}
		source := trace.NewDataDirSource(dataDir)
		return source, source.Describe(), nil
	}
	if len(args) > 0 {
		sources := make([]trace.Source, 0, len(args))
		for _, arg := range args {
			source, err := trace.SourceForPath(arg)
			if err != nil {
				return nil, nil, err
			}
			sources = append(sources, source)
		}
		source := trace.NewCombinedSource(sources)
		return source, source.Describe(), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, err
	}
	sources, err := trace.DiscoverSources(home, workspace)
	if err != nil {
		return nil, nil, err
	}
	source := trace.NewCombinedSource(sources)
	return source, source.Describe(), nil
}

func printTraceSources(descriptions []string) {
	if len(descriptions) == 0 {
		fmt.Println("Agent trace sources: none discovered")
		return
	}
	if len(descriptions) == 1 {
		fmt.Printf("Agent trace source: %s\n", descriptions[0])
		return
	}
	fmt.Println("Agent trace sources:")
	for _, desc := range descriptions {
		fmt.Printf("  %s\n", desc)
	}
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
