package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bobziuchkovski/writ"
	"github.com/fatih/color"
)

const (
	DefaultColors     = "exfxcxdxbxegedabagacad"
	TestOtherWritable = 02
	TestExecutable    = 0111
)

var LSColors []rune
var Colorize bool // FIXME: should not use a global variable here.

var FgColorMapping = map[rune][]color.Attribute{
	'a': {color.FgBlack},
	'b': {color.FgRed},
	'c': {color.FgGreen},
	'd': {color.FgYellow}, // FIXME: should be brown per ls(1).
	'e': {color.FgBlue},
	'f': {color.FgMagenta},
	'g': {color.FgCyan},
	'h': {color.FgWhite},   // FIXME: should be light grey per ls(1).
	'A': {color.FgHiWhite}, // FIXME: should be dark grey per ls(1).
	'B': {color.Bold, color.FgRed},
	'C': {color.Bold, color.FgGreen},
	'D': {color.Bold, color.FgYellow}, // FIXME: should be bold brown per ls(1).
	'E': {color.Bold, color.FgBlue},
	'F': {color.Bold, color.FgMagenta},
	'G': {color.Bold, color.FgCyan},
	'H': {color.Bold, color.FgWhite}, // FIXME: should be bold light grey per ls(1).
	'x': {},
}

var BgColorMapping = map[rune][]color.Attribute{
	'a': {color.BgBlack},
	'b': {color.BgRed},
	'c': {color.BgGreen},
	'd': {color.BgYellow}, // FIXME: should be brown per ls(1).
	'e': {color.BgBlue},
	'f': {color.BgMagenta},
	'g': {color.BgCyan},
	'h': {color.BgWhite},   // FIXME: should be light grey per ls(1).
	'A': {color.BgHiWhite}, // FIXME: should be dark grey per ls(1).
	'B': {color.Bold, color.BgRed},
	'C': {color.Bold, color.BgGreen},
	'D': {color.Bold, color.BgYellow}, // FIXME: should be bold brown per ls(1).
	'E': {color.Bold, color.BgBlue},
	'F': {color.Bold, color.BgMagenta},
	'G': {color.Bold, color.BgCyan},
	'H': {color.Bold, color.BgWhite}, // FIXME: should be bold light grey per ls(1).
	'x': {},
}

func colorized(fi os.FileInfo) string {
	m := fi.Mode()
	/*
	   Order of colors in LSCOLORS environment variable, copied from ls(1)
	      1.   directory
	      2.   symbolic link
	      3.   socket
	      4.   pipe
	      5.   executable
	      6.   block special
	      7.   character special
	      8.   executable with setuid bit set
	      9.   executable with setgid bit set
	      10.  directory writable to others, with sticky bit
	      11.  directory writable to others, without sticky bit
	*/
	var i int
	if i = 10; m&os.ModeDir&^os.ModeSticky != 0 && m&os.ModePerm&TestOtherWritable != 0 {
	} else if i = 9; m&os.ModeDir&os.ModeSticky != 0 && m&os.ModePerm&TestOtherWritable != 0 {
	} else if i = 8; m.IsRegular() && m&os.ModePerm&TestExecutable != 0 && m&os.ModeSetgid != 0 {
	} else if i = 7; m.IsRegular() && m&os.ModePerm&TestExecutable != 0 && m&os.ModeSetuid != 0 {
	} else if i = 6; m&os.ModeCharDevice != 0 {
	} else if i = 5; m&os.ModeDevice != 0 {
	} else if i = 4; m.IsRegular() && m&TestExecutable != 0 {
	} else if i = 3; m&os.ModeNamedPipe != 0 {
	} else if i = 2; m&os.ModeSocket != 0 {
	} else if i = 1; m&os.ModeSymlink != 0 {
	} else if i = 0; m.IsDir() {
	} else {
		if !Colorize {
			return fmt.Sprintf("%-12s", fi.Name())
		}
		return fmt.Sprintf("%-9s", fi.Name())
	}

	var attributes []color.Attribute
	fgCode, bgCode := LSColors[i*2], LSColors[i*2+1]
	attributes = append(attributes, FgColorMapping[fgCode]...)
	attributes = append(attributes, BgColorMapping[bgCode]...)
	return color.New(attributes...).SprintfFunc()("%-9s", fi.Name())
}

type List struct {
	HelpFlag    bool `flag:"h,help" description:"Show this help message"`
	Colorize    bool `flag:"G" description:"Enable colorize output"`
	ShowDotFile bool `flag:"a" descrition:"Include dot files (file name starts with a .)"`
}

// Entry represent a single file. It wraps os.File and os.FileInfo together.
type Entry struct {
	info os.FileInfo
}

// EntrySlice represent a list of files.
type EntrySlice []Entry

// DirEntry is dirent(5). If the positional arguments contains non-directory
// files, all of them will be considered under a virtual directory of name ""
// so the logic can be unified.
type DirEntry struct {
	name    string
	entries EntrySlice
	context *List
}

func (ls *List) Run(p writ.Path, args []string) {
	if ls.HelpFlag {
		p.Last().ExitHelp(nil)
	}

	LSColors = []rune(os.Getenv("LSCOLORS"))
	if len(LSColors) == 0 {
		LSColors = []rune(DefaultColors)
	}
	Colorize = ls.Colorize
	if len(args) == 0 {
		args = append(args, ".")
	}

	dirEntries := ls.open(args)
	var parts []string
	// There should be at least one entries, output without title
	parts = append(parts, dirEntries[0].with(ls).filtered().sorted().entries.String())
	for _, de := range dirEntries[1:] {
		s := fmt.Sprintf("%s:\n%s", de.name, de.with(ls).filtered().sorted().String())
		parts = append(parts, s)
	}
	fmt.Println(strings.Join(parts, "\n\n"))
}

func (ls *List) open(targets []string) []DirEntry {
	allDirs := []DirEntry{DirEntry{name: ""}}

	for _, target := range targets {
		// TODO: Don't exit right away. Consider a file without read permission
		// and you can't stat. This is perfect fine to list the filename but
		// not its permission bits.
		file, err := os.Open(target)
		if err != nil {
			ExitError(err)
		}
		info, err := file.Stat()
		if err != nil {
			ExitError(err)
		}
		if !info.IsDir() {
			allDirs[0].entries = append(allDirs[0].entries, Entry{info: info})
			continue
		}

		dirEntry := DirEntry{name: target}
		infos, err := file.Readdir(0)
		if err != nil {
			ExitError(err)
		}
		for _, info := range infos {
			dirEntry.entries = append(dirEntry.entries, Entry{info: info})
		}
		// TODO: by default these entries should be sort in lexical order,
		// unless there's a flag that overrides it.
		// TODO: consider -r flag and sort in reverse order.
		allDirs = append(allDirs, dirEntry)
	}
	if len(allDirs[0].entries) == 0 {
		allDirs = allDirs[1:]
	}
	sort.Sort(byDirEntryName(allDirs))
	return allDirs
}

func (de *DirEntry) with(ls *List) *DirEntry {
	de.context = ls
	return de
}

func (de *DirEntry) filtered() *DirEntry {
	newEntries := de.entries[:0]
	for _, entry := range de.entries {
		if !strings.HasPrefix(entry.info.Name(), ".") || de.context.ShowDotFile {
			newEntries = append(newEntries, entry)
		}
	}
	de.entries = newEntries
	return de
}

func (de *DirEntry) sorted() *DirEntry {
	// TODO: support more sorting method
	sort.Sort(byFilename(de.entries))
	return de
}

type byDirEntryName []DirEntry

func (xs byDirEntryName) Len() int           { return len(xs) }
func (xs byDirEntryName) Swap(i, j int)      { xs[i], xs[j] = xs[j], xs[i] }
func (xs byDirEntryName) Less(i, j int) bool { return xs[i].name < xs[j].name }

type byFilename EntrySlice

func (xs byFilename) Len() int           { return len(xs) }
func (xs byFilename) Swap(i, j int)      { xs[i], xs[j] = xs[j], xs[i] }
func (xs byFilename) Less(i, j int) bool { return xs[i].info.Name() < xs[j].info.Name() }

func (de DirEntry) String() string {
	return de.entries.String()
}

func (es EntrySlice) String() string {
	var names []string
	for _, entry := range es {
		// TODO: depending on flags and whether the output goes to terminal,
		// properly format the output.
		// TODO: 12 is the number used by OS X's ls, not sure if this is
		// universal.
		names = append(names, entry.String())
	}
	return strings.Join(names, " ")
}

func (entry Entry) String() string {
	return colorized(entry.info)
}
