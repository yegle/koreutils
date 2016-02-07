package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bobziuchkovski/writ"
)

type List struct {
	Colorize    bool `flag:"G" description:"Enable colorize output"`
	ShowDotFile bool `flag:"a" descrition:"Include dot files (file name starts with a .)"`
}

// Entry represent a single file. It wraps os.File and os.FileInfo together.
type Entry struct {
	info os.FileInfo
}

// EntrySlice represent a list of files.
type EntrySlice []Entry

// DirEntry is dirent(5). If the positional arguments contains non-directory,
// all of them will be considered under a virtual directory.
type DirEntry struct {
	name    string
	entries EntrySlice
	context *List
}

func (ls *List) Run(p writ.Path, args []string) {
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
		// TODO: ls in coreutils doesn't exit right away (consider a file
		// without read permission so you can't stat).
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
		n := entry.info.Name()
		// TODO: depending on flags and whether the output goes to terminal,
		// properly format the output.
		// TODO: 12 is the number used by OS X's ls, not sure if this is
		// universal.
		names = append(names, fmt.Sprintf("%-12s", n))
	}
	return strings.Join(names, " ")
}
