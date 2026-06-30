package main

import (
	"fmt"

	"github.com/xhd2015/less-gen/flags"
)

const statusHelp = `
Usage: agent-run status

Show agent-run home directory and basic status.
`

func runStatus(args []string) error {
	_, err := flags.Help("-h,--help", statusHelp).Parse(args)
	if err != nil {
		return err
	}
	store, err := openStore()
	if err != nil {
		return err
	}
	fmt.Printf("home: %s\n", store.Home())
	return nil
}