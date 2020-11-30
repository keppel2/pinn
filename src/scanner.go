package main

//a
import (
	"fmt"
	"io"
	//	"io/ioutil"
	//	"log"
	//	"os"
	"text/scanner"
	//	"strings"
)

func (s *scan) tokenize() {
	for {
		s.next()
		fmt.Println(s.prn())
		if s.tok == "EOF" {
			break
		}
	}

}

func (s *scan) prn() string {
	return fmt.Sprintf("%v,%v,%v\n", s.tok, s.lit, s.kind)
}

var TD = "../pinn/"

type scan struct {
	scanner.Scanner

	p    scanner.Position
	tok  string
	lit  string
	kind LitKind
}

func (s *scan) init(src io.Reader) {
	s.Init(src)
	s.p = s.Pos()
}

func (s *scan) _at() {
	s.tok += string(s.Next())
	if !tmOk(s.tok) {
		panic("")
	}
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
		if tmOk(s.tok) {
			if s.tok == "." {
				if s.Peek() == '.' {
					s.Next()
					if s.Peek() == '.' {
						s.tok = "..."
						if !tmOk(s.tok) {
							panic("")
						}
						s.Next()
						return
					}
					panic("")
				}
			}

			if s.tok == "=" {
				if s.Peek() == '=' {
					s._at()
				}
				return
			}
			if s.tok == ":" {
				if s.Peek() == '=' {
					s._at()
				}
				return
			}
			if s.tok == "<" {
				if s.Peek() == '<' {
					s._at()
				}
				if s.Peek() == '=' {
					s._at()
				}
				return
			}
			if s.tok == ">" {
				if s.Peek() == '>' {
					s._at()
				}
				if s.Peek() == '=' {
					s._at()
				}
				return
			}

			if s.tok == "&" {
				if s.Peek() == '&' {
					s._at()
				}
				return
			}
			if s.tok == "|" {
				if s.Peek() == '|' {
					s._at()
				}
				return
			}

			if s.tok == "+" {
				if s.Peek() == '+' || s.Peek() == '=' {
					s._at()
				}
				return
			}
			if s.tok == "-" {
				if s.Peek() == '-' || s.Peek() == '=' {
					s._at()
				}
				return
			}
			if s.tok == "*" {
				if s.Peek() == '=' {
					s._at()
				}
			}
			if s.tok == "/" {
				if s.Peek() == '=' {
					s._at()
				}
			}
			if s.tok == "%" {
				if s.Peek() == '=' {
					s._at()
				}
			}

			if s.tok == "!" {
				if s.Peek() == '=' {
					s._at()
				}

				return
			}
			return
		}
		panic(s.tok)
	}
}
