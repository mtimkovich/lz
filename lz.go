package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type SortBy int

const (
	NONE SortBy = iota
	TIME
	SIZE
)

type Args struct {
	t      *bool // sort by time
	s      *bool // sort by size
	r      *bool // reverse sort
	l      *bool // long
	sortby SortBy
	dirs   *[]string
}

func initArgs() (args Args) {
	kingpin.CommandLine.HelpFlag.Short('h')
	args.t = kingpin.Flag("time", "sort by modification time, oldest first").
		Short('t').Bool()
	args.s = kingpin.Flag("size", "sort by file size, smallest first").
		Short('s').Bool()
	args.r = kingpin.Flag("reverse", "reverse order while sorting").
		Short('r').Bool()
	args.l = kingpin.Flag("long", "use long listing format").
		Short('l').Bool()
	args.dirs = kingpin.Arg("directories", "show information about directory").
		Default(".").Strings()
	kingpin.Parse()

	if *args.t && *args.s {
		fmt.Println("-t and -s cannot be set at the same time.")
		os.Exit(1)
	} else if *args.t {
		args.sortby = TIME
	} else if *args.s {
		args.sortby = SIZE
	}

	return
}

func isDir(filename string) bool {
	fi, _ := os.Stat(filename)
	return fi.Mode().IsDir()
}

// Turn the string file arguments into a Files object.
func ParseFileArgs(fileArgs []string) Files {
	var fileInfos []os.FileInfo

	if len(fileArgs) == 1 && isDir(fileArgs[0]) {
		// There is only one argument, and it is a directory.
		var err error
		fileInfos, err = ioutil.ReadDir(fileArgs[0])
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Otherwise, we have a list of files/directories.
		for _, file := range fileArgs {
			fi, _ := os.Stat(file)
			fileInfos = append(fileInfos, fi)
		}
	}

	var files Files
	for _, fi := range fileInfos {
		files = append(files, NewFile(fi))
	}

	return files
}

func main() {
	args := initArgs()

	files := ParseFileArgs(*args.dirs)
	files.Sort(args.sortby, *args.r)

	// Print info
	if *args.l {
		files.PrintLong()
	} else if args.sortby != NONE {
		files.PrintSorted(args.sortby)
	} else {
		files.Print()
	}
}
