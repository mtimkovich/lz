package main

import (
	"fmt"
	"os"
	"os/user"
	"sort"
	"strconv"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/acarl005/textcol"
	humanize "github.com/dustin/go-humanize"
	"github.com/fatih/color"
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
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	blue := color.New(color.FgBlue, color.Bold).SprintFunc()

	if f.IsDir {
		return fmt.Sprintf("%v/", blue(f.FileName))
	} else if f.Executable {
		return green(f.FileName)
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
		return files[i].modTime.Before(files[j].modTime)
	})
}

func (files Files) sortBySize() {
	sort.Slice(files, func(i, j int) bool {
		return files[i].size < files[j].size
	})
}

func GetWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
}

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

func (files Files) Print() {
	var filenames []string
	for _, file := range files {
		filenames = append(filenames, file.Name())
	}

	textcol.PrintColumns(&filenames, 2)
}
