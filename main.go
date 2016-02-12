package main

import (
	"os"
	"path"

	"github.com/bobziuchkovski/writ"
	"github.com/yegle/koreutils/cmd"
)

type Runner interface {
	Run(writ.Path, []string)
}

const PROGNAME = "koreutils"

func main() {
	koreutils := &cmd.Koreutils{ProgName: PROGNAME}
	rootcmd := writ.New(PROGNAME, koreutils)

	for _, x := range cmd.AllCommands {
		rootcmd.Subcommand(x)
	}

	// Use basename, in case user run the symlink'ed command, e.g.
	// /path/to/symlinked/ls.
	if cmd := path.Base(os.Args[0]); cmd != koreutils.ProgName {
		os.Args = append([]string{PROGNAME, cmd}, os.Args[1:]...)
	}

	path, args, err := rootcmd.Decode(os.Args[1:])
	if err != nil {
		path.Last().ExitHelp(err)
	}

	switch path.String() {
	case "koreutils ls":
		koreutils.List.Run(path, args)
	default:
		koreutils.Run(path, args)
	}
}
