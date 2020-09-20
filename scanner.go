package main

//a
import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	//	"os"
	"text/scanner"

	"strings"
)

func tok(s string) {
	var got scan
	got.init(strings.NewReader(s))
	for {
		got.next()
		fmt.Println(_prn(got))
		if got.tok == "EOF" {
			break
		}
	}

}

func _prn(s scan) string {
	return fmt.Sprintf("%v,%v,%v,%v\n", s.tok, s.lit, s.kind, s.op)
}

func f() {
	rd, err := ioutil.ReadDir("pinns")
	if err != nil {
		log.Fatal(err)
	}
	for _, ofi := range rd {
		bs, _ := ioutil.ReadFile("pinns/" + ofi.Name())
		_ = bs
		src := string(bs)
		tok(src)

	}
}
func main() {
	//f()
	src := `
var x int;
123123;
"abc";`
	tok(src)
	p := new(parser)
	p.init(strings.NewReader(src))
	p.fileA()
	return
}

type scan struct {
	scanner.Scanner

	p    scanner.Position
	tok  string
	lit  string
	kind LitKind
	op   string
}

func (s *scan) init(src io.Reader) {
	s.Init(src)
}

func (s *scan) next() {
	r := s.Scan()
	s.p = s.Pos()
	s.tok = s.TokenText()
	switch r {
	case scanner.EOF:
		s.tok = "EOF"
		return
	case scanner.Int:
		s.lit = s.tok
		s.tok = "literal"
		s.kind = IntLit
		return
	case scanner.Float:
		s.lit = s.tok
		s.tok = "literal"
		s.kind = FloatLit
		return
	case scanner.String:
		s.lit = s.tok
		s.tok = "literal"
		s.kind = StringLit
		return
	case scanner.Ident:
		if !tmOk(s.tok) {
			s.lit = s.tok
			s.tok = "name"
		}
		return
	default:
		if omOk(s.tok) {
			s.lit = s.tok
			s.tok = "op"
			return
		} else if tmOk(s.tok) {
			return
		}
		panic(s.tok)
	}
}
