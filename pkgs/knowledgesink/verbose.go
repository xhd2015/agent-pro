package knowledgesink

import (
	"fmt"
	"io"
	"os"
)

func verboseWriter(opts Opts) io.Writer {
	if !opts.Verbose {
		return nil
	}
	if opts.Stderr != nil {
		return opts.Stderr
	}
	return os.Stderr
}

func verboseNotice(opts Opts, format string, args ...any) {
	w := verboseWriter(opts)
	if w == nil {
		return
	}
	fmt.Fprintf(w, "notice: "+format+"\n", args...)
}
