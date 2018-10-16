package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"sort"
	"strconv"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/acarl005/textcol"
	humanize "github.com/dustin/go-humanize"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type File struct {
	FileName   string
	Mode       os.FileMode
	Executable bool
	IsDir      bool
	modTime    time.Time
	size       uint64
	User       string
}

func NewFile(fi os.FileInfo) *File {
	file := &File{}

	file.FileName = fi.Name()
	file.Mode = fi.Mode()
	file.Executable = fi.Mode()&0111 != 0
	file.IsDir = fi.IsDir()
	file.modTime = fi.ModTime()
	file.size = uint64(fi.Size())
	lookup, _ := user.LookupId(strconv.Itoa(int(fi.Sys().(*syscall.Stat_t).Uid)))
	file.User = lookup.Username

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

func (f *File) ModTime() string {
	return humanize.Time(f.modTime)
}

func (f *File) Size() string {
	return humanize.IBytes(f.size)
}

func (f *File) Property(by SortBy) string {
	if by == TIME {
		return f.ModTime()
	} else if by == SIZE {
		return f.Size()
	}

	return ""
}

type Files []*File

func (files Files) Sort(by SortBy) {
	if by == TIME {
		files.sortByTime()
	} else if by == SIZE {
		files.sortBySize()
	}
}

func (files Files) sortByTime() {
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.After(files[j].modTime)
	})
}

func (files Files) sortBySize() {
	sort.Slice(files, func(i, j int) bool {
		return files[i].size > files[j].size
	})
}

type SortBy int

const (
	NONE SortBy = iota
	TIME
	SIZE
)

func (files Files) PrintSorted(by SortBy) {
	w := GetWriter()
	for _, file := range files {
		fmt.Fprintf(w, "%v\t%v\n", file.Property(by), file.Name())
	}
	w.Flush()
}

func (files Files) PrintLong() {
	w := GetWriter()
	for _, f := range files {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\n", f.Mode, f.User, f.Size(), f.ModTime(), f.Name())
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

type Args struct {
	t      *bool // sort by time
	s      *bool // sort by size
	r      *bool // reverse sort
	l      *bool // long
	sortby SortBy
	dir    *string
}

func initArgs() (args Args) {
	kingpin.CommandLine.HelpFlag.Short('h')
	args.t = kingpin.Flag("time", "sort by modification time, newest first").Short('t').Bool()
	args.s = kingpin.Flag("size", "sort by file size, largest first").Short('s').Bool()
	args.r = kingpin.Flag("reverse", "reverse order while sorting").Short('r').Bool()
	args.l = kingpin.Flag("long", "use long listing format").Short('l').Bool()
	args.dir = kingpin.Arg("directory", "show information about directory").Default(".").String()
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

func main() {
	args := initArgs()
	directory, err := ioutil.ReadDir(*args.dir)
	if err != nil {
		log.Fatal(err)
	}

	files := FilesInit(directory)

	files.Sort(args.sortby)
	if *args.r {
		files.Reverse()
	}

	if *args.l {
		files.PrintLong()
	} else if args.sortby != NONE {
		files.PrintSorted(args.sortby)
	} else {
		files.Print()
	}
}
