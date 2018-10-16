package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/acarl005/textcol"
	humanize "github.com/dustin/go-humanize"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type File struct {
	FileName   string
	Executable bool
	IsDir      bool
	ModTime    time.Time
	Size       uint64
}

func NewFile(fi os.FileInfo) *File {
	file := &File{}

	file.FileName = fi.Name()
	file.Executable = fi.Mode()&0111 != 0
	file.IsDir = fi.IsDir()
	file.ModTime = fi.ModTime()
	file.Size = uint64(fi.Size())

	return file
}

func (f *File) Name() string {
	if f.IsDir {
		return f.FileName + "/"
	} else if f.Executable {
		return f.FileName + "*"
	} else {
		return f.FileName
	}
}

type Files []*File

func (files Files) SortByTime() {
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})
}

func (files Files) SortBySize() {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size > files[j].Size
	})
}

type SortBy int

const (
	NONE SortBy = iota
	TIME
	SIZE
)

func (files Files) Sort(by SortBy, asc bool) {
	var property func(*File) string

	if by == TIME {
		files.SortByTime()
		property = func(f *File) string { return humanize.Time(f.ModTime) }
	} else if by == SIZE {
		files.SortBySize()
		property = func(f *File) string { return humanize.IBytes(f.Size) }
	}

	if asc {
		files.Reverse()
	}

	w := GetWriter()
	for _, file := range files {
		fmt.Fprintf(w, "%v\t%v\n", property(file), file.Name())
	}
	w.Flush()
}

func (files Files) Reverse() {
	for left, right := 0, len(files)-1; left < right; left, right = left+1, right-1 {
		files[left], files[right] = files[right], files[left]
	}
}

func FilesInit(directory []os.FileInfo) Files {
	var files Files
	for _, f := range directory {
		files = append(files, NewFile(f))
	}

	return files
}

func GetWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
}

func (files Files) Print() {
	var filenames []string
	for _, file := range files {
		filenames = append(filenames, file.Name())
	}

	textcol.PrintColumns(&filenames, 4)
}

// -l long

type Args struct {
	t   *bool // sort by time
	s   *bool // sort by size
	r   *bool // reverse sort
	dir *string
}

func initArgs() (args Args) {
	kingpin.CommandLine.HelpFlag.Short('h')
	args.t = kingpin.Flag("time", "sort by modification time, newest first").Short('t').Bool()
	args.s = kingpin.Flag("size", "sort by file size, largest first").Short('s').Bool()
	args.r = kingpin.Flag("reverse", "reverse order while sorting").Short('r').Bool()
	args.dir = kingpin.Arg("directory", "show information about directory").Default(".").String()
	kingpin.Parse()
	return
}

func main() {
	args := initArgs()
	directory, err := ioutil.ReadDir(*args.dir)
	if err != nil {
		log.Fatal(err)
	}

	if *args.t && *args.s {
		fmt.Println("-t and -s cannot be set at the same time.")
		os.Exit(1)
	}
	var sortby SortBy = NONE
	if *args.t {
		sortby = TIME
	} else if *args.s {
		sortby = SIZE
	}

	files := FilesInit(directory)

	if sortby == NONE {
		files.Print()
	} else {
		files.Sort(sortby, *args.r)
	}
}
