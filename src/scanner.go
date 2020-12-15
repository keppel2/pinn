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

type tlt []*token

func (t tlt) last() *token {
	if len(t) == 0 {
		return nil
	}
	return t[len(t)-1]
}

func (t tlt) len() int {
	return len(t)
}

type scan struct {
	ss     scanner.Scanner
	tks    []*token
	cursor int
	qmarks []tlt
}

func (s *scan) qmark() *token {
	return s.qmarks[len(s.qmarks)-1].last()
}

func (s *scan) qmpush() {
	s.qmarks[len(s.qmarks)-1] = append(s.qmarks[len(s.qmarks)-1], s.ct())
}

func (s *scan) qmpop() {
	s.qmarks[len(s.qmarks)-1] = s.qmarks[len(s.qmarks)-1][0 : s.qmarks[len(s.qmarks)-1].len()-1]

}

func (s *scan) ct() *token {
	return s.tks[s.cursor]
}

func (s *scan) init(src io.Reader) {
	s.qmarks = make([]tlt, 1)
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
				s.qmpush()
			}
			if s.ct().tok == ";" {
				s.qmarks = make([]tlt, 1)
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

						if len(s.qmarks[len(s.qmarks)-1]) > 1 {
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
