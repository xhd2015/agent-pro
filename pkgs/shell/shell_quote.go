package shell

import "github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"

// ShellQuote wraps a string in POSIX-safe shell quoting.
func ShellQuote(s string) string {
	return ptywrap.ShellQuote(s)
}