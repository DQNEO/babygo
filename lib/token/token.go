package token

import "github.com/DQNEO/babygo/lib/fmt"

type Token string
type Pos int

// Kind
var INT Token = "INT"
var STRING Token = "STRING"

var NoPos Pos = 0

// Token
var ADD Token = "+"
var SUB Token = "-"
var AND Token = "&"

func (tok Token) String() string {
	return string(tok)
}

type File struct {
	Name  string // absolute path
	Base  int
	Lines []Pos // index is line number - 1. value is the position of the start of each line
	Size  int
}

type FileSet struct {
	Base  int
	Files []*File
}

func NewFileSet() *FileSet {
	return &FileSet{
		Base: 1,
	}
}

func (fs *FileSet) AddFile(filename string, base int, size int) *File {
	f := &File{
		Name: filename,
		Size: size,
	}
	if base < 0 {
		base = fs.Base
	}
	f.Base = base
	fs.Files = append(fs.Files, f)
	fs.Base += size + 1
	return f
}

type Position struct {
	Filename string // filename, if any
	Offset   int    // offset, starting at 0
	Line     int    // line number, starting at 1
	Column   int    // column number, starting at 1 (byte count)
}

func (p *Position) String() string {
	return fmt.Sprintf("%s:%d", p.Filename, p.Line)
}
func (fs *FileSet) Position(pos Pos) *Position {
	var currentFile *File
	if len(fs.Files) > 0 {
		currentFile = fs.Files[0]
	}
	for _, nextFile := range fs.Files[1:] {
		if int(pos) < nextFile.Base {
			break
		}
		currentFile = nextFile
	}

	// debug:
	//	fmt.Fprintf(os.Stderr, "[token.Position] currentFile=%s, firstPos=%d\n", currentFile.Name, int(currentFile.Lines[0]))

	var currentLineNo int = 1 // Line starts from 1
	for id, nextLinePos := range currentFile.Lines {
		if pos < nextLinePos {
			break
		}
		currentLineNo = id + 1
	}

	return &Position{
		Filename: currentFile.Name,
		Offset:   0,
		Line:     currentLineNo,
		Column:   0,
	}
}
