package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bobziuchkovski/writ"
	"github.com/kardianos/osext"
)

type Koreutils struct {
	HelpFlag  bool `flag:"h,help" description:"Help of koreutils"`
	DoInstall bool `flag:"install" description:"Install commands by create symlinks"`
	List      List `command:"ls" alias:"list" description:"List directory content"`
}

func (k *Koreutils) Run(p writ.Path, args []string) {
	if len(args) != 0 {
		p.Last().ExitHelp(ExitMessage{fmt.Sprintf("unrecognized subcommand: `%s`", args[0])})
	}

	if k.HelpFlag {
		p.Last().ExitHelp(nil)
	} else if k.DoInstall {
		executable, err := osext.Executable()
		if err != nil {
			ExitError(err)
		}
		dir := filepath.Dir(executable)
		err = k.Install(dir)
		if err != nil {
			ExitError(err)
		}
	} else {
		p.Last().ExitHelp(ExitMessage{"no flag given, show help"})
	}
}

func (*Koreutils) Install(dir string) error {
	for cmd := range AllCommands {
		err := os.Symlink(PROGNAME, filepath.Join(dir, cmd))
		if err != nil {
			return err
		}
	}
	return nil
}
