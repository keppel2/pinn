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
	for k, tk := range s.tks {
		fmt.Println(k, tk.prn())
	}
}

func (t *token) prn() string {
	return fmt.Sprintf("%v,%v,%v,%v\n", t.tok, t.lit, t.lk, t.colons)
}

var TD = "../pinn/"

type token struct {
	tok    string
	lit    string
	lk     LitKind
	p      scanner.Position
	colons int
}

type scan struct {
	ss     scanner.Scanner
	tks    []*token
	cursor int
	qmarks []*token
}

func (s *scan) qmark() *token {
	if len(s.qmarks) == 0 {
		return nil
	}
	return s.qmarks[len(s.qmarks)-1]
}

func (s *scan) qmpush() {
	s.qmarks = append(s.qmarks, s.ct())
}
func (s *scan) qmpop() {
	s.qmarks = s.qmarks[0 : len(s.qmarks)-1]
}

func (s *scan) ct() *token {
	return s.tks[s.cursor]
}

func (s *scan) init(src io.Reader) {
	s.ss.Init(src)
	s.cursor = -1
	for {
		s.next()
		if s.ct().tok == "EOF" {
			break
		}
	}
	s.cursor = -1
}

func (s *scan) _at() {
	s.ct().tok += string(s.ss.Next())
	if !tmOk(s.ct().tok) {
		panic("")
	}
}

func (s *scan) next() {
	r := s.ss.Scan()
	t := new(token)
	s.cursor++
	s.tks = append(s.tks, t)
	t.p = s.ss.Pos()
	t.tok = s.ss.TokenText()
	switch r {
	case scanner.EOF:
		s.ct().tok = "EOF"
		return
	case scanner.Int:
		s.ct().lit = s.ct().tok
		s.ct().tok = "literal"
		s.ct().lk = IntLit
		return
	case scanner.Float:
		s.ct().lit = s.ct().tok
		s.ct().tok = "literal"
		s.ct().lk = FloatLit
		return
	case scanner.String:
		s.ct().lit = s.ct().tok
		s.ct().tok = "literal"
		s.ct().lk = StringLit
		return
	case scanner.Ident:
		if !tmOk(s.ct().tok) {
			s.ct().lit = s.ct().tok
			s.ct().tok = "name"
		}
		return
	default:
		if tmOk(s.ct().tok) {
			if s.ct().tok == "." {
				if s.ss.Peek() == '.' {
					s.ss.Next()
					if s.ss.Peek() == '.' {
						s.ct().tok = "..."
						if !tmOk(s.ct().tok) {
							panic("")
						}
						s.ss.Next()
						return
					}
					panic("")
				}
			}

			if s.ct().tok == "?" {
				s.qmarks = append(s.qmarks, s.ct())
			}
			if s.ct().tok == ";" {
				s.qmarks = nil
			}

			if s.ct().tok == "=" {
				if s.ss.Peek() == '=' {
					s._at()
				}
				return
			}
			if s.ct().tok == ":" {
				if s.ss.Peek() == '=' {
					s._at()
				} else {
					if s.qmark() != nil {
						s.qmark().colons++

						if len(s.qmarks) > 1 {
							s.qmpop()
						}
					}
				}
				return
			}
			if s.ct().tok == "<" {
				if s.ss.Peek() == '<' {
					s._at()
				}
				if s.ss.Peek() == '=' {
					s._at()
				}
				return
			}
			if s.ct().tok == ">" {
				if s.ss.Peek() == '>' {
					s._at()
				}
				if s.ss.Peek() == '=' {
					s._at()
				}
				return
			}

			if s.ct().tok == "&" {
				if s.ss.Peek() == '&' {
					s._at()
				}
				return
			}
			if s.ct().tok == "|" {
				if s.ss.Peek() == '|' {
					s._at()
				}
				return
			}

			if s.ct().tok == "+" {
				if s.ss.Peek() == '+' || s.ss.Peek() == '=' {
					s._at()
				}
				return
			}
			if s.ct().tok == "-" {
				if s.ss.Peek() == '-' || s.ss.Peek() == '=' {
					s._at()
				}
				return
			}
			if s.ct().tok == "*" {
				if s.ss.Peek() == '=' {
					s._at()
				}
			}
			if s.ct().tok == "/" {
				if s.ss.Peek() == '=' {
					s._at()
				}
			}
			if s.ct().tok == "%" {
				if s.ss.Peek() == '=' {
					s._at()
				}
			}

			if s.ct().tok == "!" {
				if s.ss.Peek() == '=' {
					s._at()
				}

				return
			}
			return
		}
		panic(s.ct().tok)
	}
}
