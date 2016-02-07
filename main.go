package main

import (
	"os"

	"github.com/bobziuchkovski/writ"
)

type Runner interface {
	Run(writ.Path, []string)
}

const PROGNAME = "koreutils"

var AllCommands = map[string]Runner{
	"ls": &List{},
}

func main() {
	koreutils := &Koreutils{}
	rootcmd := writ.New("koreutils", koreutils)
	//rootcmd.Help.Usage = "Usage: koreutils [--help] [--install] COMMAND [OPTION]... [ARG]..."

	for x := range AllCommands {
		// Register subcommands
		rootcmd.Subcommand(x)
	}

	if os.Args[0] != PROGNAME {
		os.Args = append([]string{PROGNAME}, os.Args...)
	}

	path, args, err := rootcmd.Decode(os.Args[1:])
	if err != nil {
		path.Last().ExitHelp(err)
	}

	var runner Runner

	if len(path) > 1 {
		runner = AllCommands[path[1].String()]
	}

	if runner == nil {
		runner = koreutils
	}

	runner.Run(path, args)
}
