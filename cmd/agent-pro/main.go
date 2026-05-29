package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tidwall/jsonc"
	"github.com/xhd2015/agent-pro/agent/opencode/commands"
	"github.com/xhd2015/agent-pro/agent/opencode/config"
	"github.com/xhd2015/agent-pro/agent/opencode/permissions"
	"github.com/xhd2015/agent-pro/agent/opencode/plugins"
	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: agent-pro <command> [ARGS]

Commands:
  opencode          manage opencode hooks and permissions

Run agent-pro <command> --help for command-specific options.
`

const opencodeHelp = `
Usage: agent-pro opencode <command> [ARGS]

Commands:
  commands          list opencode slash commands
  permissions       manage opencode permissions
  plugins           manage opencode plugins

Run agent-pro opencode <command> --help for command-specific options.
`

func main() {
	if err := handle(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "agent-pro: %v\n", err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return nil
	}

	switch args[0] {
	case "opencode":
		return handleOpenCode(args[1:])
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func handleOpenCode(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(opencodeHelp, "\n"))
		return nil
	}

	switch args[0] {
	case "permissions":
		return handlePermissions(args[1:])
	case "plugins":
		return handlePlugins(args[1:])
	case "commands":
		return handleCommands(args[1:])
	default:
		return fmt.Errorf("unknown opencode command: %s", args[0])
	}
}

const commandsHelp = `
Usage: agent-pro opencode commands <command> [ARGS]

Commands:
  list [--dir DIR]  list slash commands (global and local)

Options:
  --dir <dir>       project directory for local commands (default: current directory)
  -h,--help         show help
`

func handleCommands(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(commandsHelp, "\n"))
		return nil
	}

	switch args[0] {
	case "list":
		return handleCommandsList(args[1:])
	default:
		return fmt.Errorf("unknown commands command: %s", args[0])
	}
}

func handleCommandsList(args []string) error {
	var dirFlag *string
	_, err := flags.String("--dir", &dirFlag).
		Help("-h,--help", `
Usage: agent-pro opencode commands list [--dir DIR]

List slash commands from opencode config and markdown files.

Options:
  --dir <dir>   project directory (default: current directory)
  -h,--help     show help
`).
		Parse(args)
	if err != nil {
		return err
	}

	dir := "."
	if dirFlag != nil && strings.TrimSpace(*dirFlag) != "" {
		dir = strings.TrimSpace(*dirFlag)
	}
	dir, err = filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home dir: %w", err)
	}

	printCommandsLocation("Global", filepath.Join(home, ".config", "opencode"), dir)
	printCommandsLocation("Local", filepath.Join(dir, ".opencode"), dir)

	return nil
}

func printCommandsLocation(label, opencodeDir, baseDir string) {
	fmt.Printf("%s (%s):\n", label, opencodeDir)
	fmt.Printf("  sources: opencode.jsonc command field, {command,commands}/**/*.md\n")

	cmds, err := commands.List(opencodeDir)
	if err != nil {
		fmt.Printf("  (error: %v)\n\n", err)
		return
	}

	if len(cmds) == 0 {
		fmt.Println("  No commands found")
		return
	}

	fmt.Printf("  Found %d command(s):\n", len(cmds))
	for _, c := range cmds {
		if c.Path != "" {
			rel, _ := filepath.Rel(baseDir, c.Path)
			fmt.Printf("    %-35s %s\n      file: %s  source: %s\n", c.Name, c.Description, rel, c.Source)
		} else {
			fmt.Printf("    %-35s %s  source: %s\n", c.Name, c.Description, c.Source)
		}
	}
	fmt.Println()
}

const pluginsHelp = `
Usage: agent-pro opencode plugins <command> [ARGS]

Commands:
  list [--dir DIR]    list installed plugins (global and local)
  add <plugin.ts>     install a plugin file to local .opencode/plugins/
  add --global <...>  install a plugin file to global ~/.config/opencode/plugins/

Options:
  --dir <dir>         project directory for local plugins (default: current directory)
  -h,--help           show help
`

func handlePlugins(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(pluginsHelp, "\n"))
		return nil
	}

	switch args[0] {
	case "list":
		return handlePluginsList(args[1:])
	case "add":
		return handlePluginsAdd(args[1:])
	default:
		return fmt.Errorf("unknown plugins command: %s", args[0])
	}
}

func handlePluginsList(args []string) error {
	var dirFlag *string
	_, err := flags.String("--dir", &dirFlag).
		Help("-h,--help", `
Usage: agent-pro opencode plugins list [--dir DIR]

List installed plugins from auto-discovered plugin directories.

Options:
  --dir <dir>   project directory (default: current directory)
  -h,--help     show help
`).
		Parse(args)
	if err != nil {
		return err
	}

	dir := "."
	if dirFlag != nil && strings.TrimSpace(*dirFlag) != "" {
		dir = strings.TrimSpace(*dirFlag)
	}
	dir, err = filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home dir: %w", err)
	}

	printPluginLocation("Global", filepath.Join(home, ".config", "opencode"), dir)
	printPluginLocation("Local", filepath.Join(dir, ".opencode"), dir)

	return nil
}

func printPluginLocation(label, opencodeDir, baseDir string) {
	fmt.Printf("%s (%s):\n", label, opencodeDir)
	fmt.Printf("  scanned: plugins/*.ts, plugins/*.js, plugin/*.ts, plugin/*.js\n")

	result, err := plugins.List(opencodeDir)
	if err != nil {
		fmt.Printf("  (error: %v)\n\n", err)
		return
	}

	if len(result) == 0 {
		fmt.Println("  No plugins found")
		return
	}

	fmt.Printf("  Found %d plugin(s):\n", len(result))
	for _, p := range result {
		rel, _ := filepath.Rel(baseDir, p.Path)
		if rel == "" {
			rel = p.Path
		}
		fmt.Printf("    %s  (%s)\n", p.Name, rel)
	}
	fmt.Println()
}

func handlePluginsAdd(args []string) error {
	var globalFlag bool
	remaining, err := flags.Bool("--global", &globalFlag).
		Help("-h,--help", `
Usage: agent-pro opencode plugins add <plugin.ts> [--global]

Install a plugin file to the opencode auto-discovered plugins directory.
Without --global, installs to .opencode/plugins/ in the current project.
With --global, installs to ~/.config/opencode/plugins/.

Options:
  --global   install to global plugins directory
  -h,--help  show help
`).
		Parse(args)
	if err != nil {
		return err
	}

	if len(remaining) == 0 {
		return fmt.Errorf("plugin file path is required")
	}
	pluginPath := remaining[0]

	if globalFlag {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("find home dir: %w", err)
		}
		destDir := filepath.Join(home, ".config", "opencode")
		dstPath, err := plugins.InstallToDir(destDir, pluginPath)
		if err != nil {
			return err
		}
		fmt.Printf("Installed plugin: %s\n", dstPath)
		return nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current dir: %w", err)
	}
	destDir := filepath.Join(cwd, ".opencode")
	dstPath, err := plugins.InstallToDir(destDir, pluginPath)
	if err != nil {
		return err
	}
	fmt.Printf("Installed plugin: %s\n", dstPath)
	return nil
}

const permissionsHelp = `
Usage: agent-pro opencode permissions <command> [ARGS]

Commands:
  list [--dir DIR]  list configured permission (deny) rules (global and local)

Options:
  --dir <dir>       project directory for local permissions (default: current directory)
  -h,--help         show help
`

func handlePermissions(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(permissionsHelp, "\n"))
		return nil
	}

	switch args[0] {
	case "list":
		return handlePermissionsList(args[1:])
	default:
		return fmt.Errorf("unknown permissions command: %s", args[0])
	}
}

func handlePermissionsList(args []string) error {
	var dirFlag *string
	_, err := flags.String("--dir", &dirFlag).
		Help("-h,--help", `
Usage: agent-pro opencode permissions list [--dir DIR]

List configured permission (deny) rules from the opencode config.

Options:
  --dir <dir>   project directory (default: current directory)
  -h,--help     show help
`).
		Parse(args)
	if err != nil {
		return err
	}

	dir := "."
	if dirFlag != nil && strings.TrimSpace(*dirFlag) != "" {
		dir = strings.TrimSpace(*dirFlag)
	}
	dir, err = filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find home dir: %w", err)
	}

	printPermissionLocation("Global", filepath.Join(home, ".config", "opencode"))
	printPermissionLocation("Local", filepath.Join(dir, ".opencode"))

	return nil
}

func printPermissionLocation(label, opencodeDir string) {
	fmt.Printf("%s (%s):\n", label, opencodeDir)
	fmt.Printf("  scanned: opencode.jsonc, opencode.json\n")

	cfg, err := readConfigDir(opencodeDir)
	if err != nil {
		fmt.Printf("  (error: %v)\n\n", err)
		return
	}

	printPermissionRules(cfg)
}

func readConfigDir(opencodeDir string) (*config.Config, error) {
	var configPath string
	var raw []byte
	for _, name := range []string{"opencode.jsonc", "opencode.json"} {
		p := filepath.Join(opencodeDir, name)
		data, err := os.ReadFile(p)
		if err == nil {
			configPath = p
			raw = data
			break
		}
	}
	if configPath == "" {
		configPath = filepath.Join(opencodeDir, "opencode.json")
		return &config.Config{Path: configPath, Data: config.Data{}}, nil
	}

	cleaned := jsonc.ToJSON(raw)
	var data config.Data
	if err := json.Unmarshal(cleaned, &data); err != nil {
		return nil, fmt.Errorf("parse %s: %w", configPath, err)
	}
	return &config.Config{Path: configPath, Data: data}, nil
}

func printPermissionRules(cfg *config.Config) {
	bashPerm := permissions.GetBash(cfg.Data)
	if bashPerm == nil {
		fmt.Println("  No deny rules configured")
		return
	}

	obj, ok := bashPerm.(map[string]interface{})
	if !ok {
		fmt.Println("  No deny rules configured")
		return
	}

	keys := config.SortedKeys(obj)
	if len(keys) == 0 {
		fmt.Println("  No deny rules configured")
		return
	}

	fmt.Printf("  Found %d rule(s):\n", len(keys))
	for _, key := range keys {
		action, _ := obj[key].(string)
		fmt.Printf("    %-40s %s\n", key, action)
	}
	fmt.Println()
}
