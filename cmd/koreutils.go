package cmd

import (
	"os"
	"path/filepath"

	"github.com/bobziuchkovski/writ"
	"github.com/kardianos/osext"
)

type Koreutils struct {
	ProgName  string
	HelpFlag  bool `flag:"h,help" description:"Show this help message"`
	DoInstall bool `flag:"install" description:"Install commands by create symlinks"`
	List      List `command:"ls" alias:"list" description:"List directory content"`
}

var AllCommands = []string{
	"ls",
}

func (k *Koreutils) Run(p writ.Path, args []string) {
	if len(args) != 0 {
		p.Last().ExitHelp(ExitMessage{"unrecognized subcommand: " + args[0]})
	}

	if k.HelpFlag {
		p.Last().ExitHelp(nil)
	}

	if !k.DoInstall {
		p.Last().ExitHelp(ExitMessage{"no flag given, show help"})
	}
	executable, err := osext.Executable()
	if err != nil {
		ExitError(err)
	}
	dir := filepath.Dir(executable)
	err = k.Install(dir)
	if err != nil {
		ExitError(err)
	}
}

func (k *Koreutils) Install(dir string) error {
	for _, cmd := range AllCommands {
		err := os.Symlink(k.ProgName, filepath.Join(dir, cmd))
		if err != nil {
			return err
		}
	}
	return nil
}
