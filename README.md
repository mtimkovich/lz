# lz

`ls` but with simplified options and more accessible output.

## Installation

```
go get github.com/mtimkovich/lz
```

## Usage

```
usage: lz [<flags>] [<directory>]

Flags:
  -h, --help     Show context-sensitive help (also try --help-long and --help-man).
  -t, --time     sort by modification time, oldest first
  -s, --size     sort by file size, smallest first
  -r, --reverse  reverse order while sorting
  -l, --long     use long listing format

Args:
  [<directory>]  show information about directory
```
